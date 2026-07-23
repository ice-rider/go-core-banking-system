package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/transaction/domain"
	"go-core-banking-system/internal/transaction/mocks"
)

func TestTransfer_ValidTransfer(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	var createdTx *domain.Transaction
	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		createdTx = tx
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
	assert.Equal(t, "acc-from", result.FromAccountID)
	assert.Equal(t, "acc-to", result.ToAccountID)
	assert.Equal(t, int64(500), result.Amount)
	assert.Equal(t, "key-1", result.IdempotencyKey)
	assert.NotNil(t, createdTx)
}

func TestTransfer_SameAccount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "same-id",
		ToAccountID:    "same-id",
		Amount:         100,
		IdempotencyKey: "key-1",
	})

	require.ErrorIs(t, err, domain.ErrSameAccount)
}

func TestTransfer_NegativeAmount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         -50,
		IdempotencyKey: "key-1",
	})

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransfer_ZeroAmount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         0,
		IdempotencyKey: "key-1",
	})

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransfer_EmptyIdempotencyKey(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "",
	})

	require.ErrorIs(t, err, domain.ErrIdempotencyKeyRequired)
}

func TestTransfer_Idempotent(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	existingTx := &domain.Transaction{
		ID:             "existing-tx",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         200,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "key-1",
	}

	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return existingTx, nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         200,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, existingTx, result)
}

func TestTransfer_UniqueViolation(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	existingTx := &domain.Transaction{
		ID:             "existing-tx",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         200,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "key-1",
	}

	callCount := 0
	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		callCount++
		if callCount == 1 {
			return nil, domain.ErrTransactionNotFound
		}
		return existingTx, nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return &pgconn.PgError{Code: "23505", Message: "duplicate key violates unique constraint"}
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         200,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, existingTx, result)
}

func TestExecuteSaga_Success(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         300,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestExecuteSaga_ReserveError(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	reserveErr := errors.New("reserve failed")
	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return reserveErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}

	var updatedStatus domain.TransactionStatus
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		updatedStatus = status
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         300,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, reserveErr, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TxStatusFailed, updatedStatus)
}

func TestExecuteSaga_CreditErrorCompensated(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	creditErr := errors.New("credit failed")
	accountCli.CreditFunc = func(ctx context.Context, id string, amount int64) error {
		return creditErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}

	var updatedStatus domain.TransactionStatus
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		updatedStatus = status
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         300,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, creditErr, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TxStatusFailed, updatedStatus)
}

func TestExecuteSaga_CommitReservationErrorCompensated(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	commitErr := errors.New("commit failed")
	accountCli.CommitReservationFunc = func(ctx context.Context, id string, amount int64) error {
		return commitErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}

	var updatedStatus domain.TransactionStatus
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		updatedStatus = status
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         300,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, commitErr, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.TxStatusFailed, updatedStatus)
}

func TestGetByID_ExistingTransaction(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	expectedTx := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "key-1",
	}

	repo.GetByIDFunc = func(ctx context.Context, id string) (*domain.Transaction, error) {
		return expectedTx, nil
	}

	result, err := svc.GetByID(context.Background(), "tx-1")

	require.NoError(t, err)
	assert.Equal(t, expectedTx, result)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.GetByIDFunc = func(ctx context.Context, id string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	result, err := svc.GetByID(context.Background(), "nonexistent-id")

	require.ErrorIs(t, err, domain.ErrTransactionNotFound)
	assert.Nil(t, result)
}

func TestTransfer_PublishEventOnSuccess(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	var publishedEvent *domain.TransactionEvent
	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		publishedEvent = event
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
	require.NotNil(t, publishedEvent)
	assert.Equal(t, domain.EventTransactionCompleted, publishedEvent.Type)
	assert.Equal(t, "acc-from", publishedEvent.FromAccountID)
	assert.Equal(t, "acc-to", publishedEvent.ToAccountID)
	assert.Equal(t, int64(500), publishedEvent.Amount)
}

func TestTransfer_PublishEventOnFailure(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return errors.New("insufficient funds")
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	var publishedEvent *domain.TransactionEvent
	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		publishedEvent = event
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	require.NotNil(t, publishedEvent)
	assert.Equal(t, domain.EventTransactionFailed, publishedEvent.Type)
}

func TestTransfer_NilPublisher_NoError(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	svc := NewTransactionService(repo, accountCli, nil)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestIsUniqueViolation_NilError(t *testing.T) {
	assert.False(t, isUniqueViolation(nil))
}

func TestIsUniqueViolation_NonPgError(t *testing.T) {
	assert.False(t, isUniqueViolation(errors.New("some error")))
}

func TestIsUniqueViolation_WrongCode(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23503"}
	assert.False(t, isUniqueViolation(pgErr))
}

func TestUpdateStatus_Error(t *testing.T) {
	repo := &mocks.MockTransactionRepository{
		UpdateStatusFunc: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			return errors.New("db error")
		},
	}
	svc := NewTransactionService(repo, nil, nil).(*transactionService)
	assert.NotPanics(t, func() {
		svc.updateStatus(context.Background(), "tx-1", domain.TxStatusFailed)
	})
}

func TestCompensateCancelReservation_Error(t *testing.T) {
	accountCli := &mocks.MockAccountClient{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return errors.New("cancel failed")
		},
	}
	svc := NewTransactionService(nil, accountCli, nil).(*transactionService)
	tx := &domain.Transaction{ID: "tx-1", FromAccountID: "acc-1", Amount: 100}
	assert.NotPanics(t, func() {
		svc.compensateCancelReservation(context.Background(), tx)
	})
}

func TestCompensateFull_DebitError(t *testing.T) {
	accountCli := &mocks.MockAccountClient{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return errors.New("debit failed")
		},
	}
	svc := NewTransactionService(nil, accountCli, nil).(*transactionService)
	tx := &domain.Transaction{ID: "tx-1", FromAccountID: "acc-1", ToAccountID: "acc-2", Amount: 100}
	assert.NotPanics(t, func() {
		svc.compensateFull(context.Background(), tx)
	})
}

func TestTransfer_UpdateStatusError(t *testing.T) {
	repo := &mocks.MockTransactionRepository{
		CreateFunc: func(ctx context.Context, tx *domain.Transaction) error {
			return nil
		},
		UpdateStatusFunc: func(ctx context.Context, id string, status domain.TransactionStatus) error {
			return errors.New("update failed")
		},
	}
	accountCli := &mocks.MockAccountClient{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return errors.New("reserve failed")
		},
	}
	svc := NewTransactionService(repo, accountCli, nil)
	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-1",
	})
	require.Error(t, err)
}

func TestTransfer_PublishEventError_Ignored(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		return errors.New("rabbitmq connection lost")
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestTransfer_FromEqualsTo(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	_, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-1",
		Amount:         100,
		IdempotencyKey: "key-1",
	})

	require.ErrorIs(t, err, domain.ErrSameAccount)
}

func TestTransfer_EmptyFromAccount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
	assert.Equal(t, "", result.FromAccountID)
}

func TestTransfer_EmptyToAccount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "",
		Amount:         100,
		IdempotencyKey: "key-1",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
	assert.Equal(t, "", result.ToAccountID)
}

func TestTransfer_MaxInt64Amount(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         math.MaxInt64,
		IdempotencyKey: "key-max",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(math.MaxInt64), result.Amount)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestTransfer_VeryLongIdempotencyKey(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	longKey := strings.Repeat("k", 10000)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: longKey,
	})

	require.NoError(t, err)
	assert.Equal(t, longKey, result.IdempotencyKey)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestTransfer_DuplicateKey_DifferentParams(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	existingTx := &domain.Transaction{
		ID:             "existing-tx",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         200,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "key-dup",
	}

	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return existingTx, nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from-other",
		ToAccountID:    "acc-to-other",
		Amount:         999,
		IdempotencyKey: "key-dup",
	})

	require.NoError(t, err)
	assert.Equal(t, existingTx, result)
	assert.Equal(t, int64(200), result.Amount)
	assert.Equal(t, "acc-from", result.FromAccountID)
	assert.Equal(t, "acc-to", result.ToAccountID)
}

func TestTransfer_DuplicateKey_SameParams(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	existingTx := &domain.Transaction{
		ID:             "existing-tx",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "key-same",
	}

	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return existingTx, nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-same",
	})

	require.NoError(t, err)
	assert.Equal(t, existingTx, result)
	assert.Equal(t, "existing-tx", result.ID)
}

func TestTransfer_IdempotencyKey_ReusedAfterFailure(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	failedTx := &domain.Transaction{
		ID:             "failed-tx",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		Status:         domain.TxStatusFailed,
		IdempotencyKey: "key-retry",
	}

	callCount := 0
	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		callCount++
		if callCount == 1 {
			return nil, domain.ErrTransactionNotFound
		}
		return failedTx, nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}

	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return errors.New("reserve failed")
	}

	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-retry",
	})

	require.Error(t, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)

	result2, err2 := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-retry",
	})

	require.NoError(t, err2)
	assert.Equal(t, failedTx, result2)
	assert.Equal(t, domain.TxStatusFailed, result2.Status)
}

func TestTransfer_SagaReserveFails_CompensateNothing(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	reserveErr := errors.New("insufficient funds")
	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return reserveErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, reserveErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.False(t, cancelCalled.Load(), "CancelReservation should not be called when Reserve fails")
}

func TestTransfer_SagaCreditFails_CancelReserveSucceeds(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	creditErr := errors.New("credit failed")
	accountCli.CreditFunc = func(ctx context.Context, id string, amount int64) error {
		return creditErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, creditErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.True(t, cancelCalled.Load(), "CancelReservation should be called when Credit fails")
}

func TestTransfer_SagaCreditFails_CancelReserveFails(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	creditErr := errors.New("credit failed")
	accountCli.CreditFunc = func(ctx context.Context, id string, amount int64) error {
		return creditErr
	}

	cancelErr := errors.New("cancel reservation also failed")
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		return cancelErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, creditErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
}

func TestTransfer_SagaCommitFails_DebitSucceeds(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	commitErr := errors.New("commit failed")
	accountCli.CommitReservationFunc = func(ctx context.Context, id string, amount int64) error {
		return commitErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	var debitCalled atomic.Bool
	accountCli.DebitFunc = func(ctx context.Context, id string, amount int64) error {
		debitCalled.Store(true)
		return nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, commitErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.True(t, cancelCalled.Load(), "CancelReservation should be called during full compensation")
	assert.True(t, debitCalled.Load(), "Debit should be called during full compensation")
}

func TestTransfer_SagaCommitFails_DebitFails(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	commitErr := errors.New("commit failed")
	accountCli.CommitReservationFunc = func(ctx context.Context, id string, amount int64) error {
		return commitErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	debitErr := errors.New("debit failed")
	accountCli.DebitFunc = func(ctx context.Context, id string, amount int64) error {
		return debitErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, commitErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.True(t, cancelCalled.Load(), "CancelReservation should still be called even if Debit fails")
}

func TestTransfer_Saga_AllStepsFail(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	reserveErr := errors.New("reserve failed")
	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return reserveErr
	}

	var creditCalled atomic.Bool
	accountCli.CreditFunc = func(ctx context.Context, id string, amount int64) error {
		creditCalled.Store(true)
		return errors.New("credit failed")
	}

	var commitCalled atomic.Bool
	accountCli.CommitReservationFunc = func(ctx context.Context, id string, amount int64) error {
		commitCalled.Store(true)
		return errors.New("commit failed")
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	})

	require.Error(t, err)
	assert.Equal(t, reserveErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.False(t, creditCalled.Load(), "Credit should not be called when Reserve fails")
	assert.False(t, commitCalled.Load(), "CommitReservation should not be called when Reserve fails")
}

func TestTransfer_Concurrent_SameAccounts(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}
	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)
	results := make([]*domain.Transaction, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.Transfer(context.Background(), domain.TransferInput{
				FromAccountID:  "acc-from",
				ToAccountID:    "acc-to",
				Amount:         int64(100 + idx),
				IdempotencyKey: "key-concurrent-" + string(rune('A'+idx)),
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errs[i], "goroutine %d failed", i)
		assert.Equal(t, domain.TxStatusCompleted, results[i].Status, "goroutine %d wrong status", i)
		assert.Equal(t, "acc-from", results[i].FromAccountID)
		assert.Equal(t, "acc-to", results[i].ToAccountID)
	}
}

func TestTransfer_Concurrent_SameIdempotencyKey(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}
	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	const numGoroutines = 5
	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)
	results := make([]*domain.Transaction, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.Transfer(context.Background(), domain.TransferInput{
				FromAccountID:  "acc-from",
				ToAccountID:    "acc-to",
				Amount:         500,
				IdempotencyKey: "key-shared",
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errs[i], "goroutine %d failed", i)
		assert.Equal(t, domain.TxStatusCompleted, results[i].Status)
	}
}

func TestTransfer_Concurrent_DifferentAccounts(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}
	repo.GetByIdempotencyKeyFunc = func(ctx context.Context, key string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	const numGoroutines = 8
	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)
	results := make([]*domain.Transaction, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.Transfer(context.Background(), domain.TransferInput{
				FromAccountID:  "acc-from-" + string(rune('A'+idx)),
				ToAccountID:    "acc-to-" + string(rune('A'+idx)),
				Amount:         int64(100 + idx),
				IdempotencyKey: "key-diff-" + string(rune('A'+idx)),
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		require.NoError(t, errs[i], "goroutine %d failed", i)
		assert.Equal(t, domain.TxStatusCompleted, results[i].Status)
	}
}

func TestTransfer_Success_PublishesCompletedEvent(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	var publishedEvent *domain.TransactionEvent
	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		publishedEvent = event
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         777,
		IdempotencyKey: "key-pub",
	})

	require.NoError(t, err)
	require.NotNil(t, publishedEvent)
	assert.Equal(t, domain.EventTransactionCompleted, publishedEvent.Type)
	assert.Equal(t, domain.TxStatusCompleted, publishedEvent.Status)
	assert.Equal(t, result.ID, publishedEvent.TransactionID)
	assert.Equal(t, "acc-from", publishedEvent.FromAccountID)
	assert.Equal(t, "acc-to", publishedEvent.ToAccountID)
	assert.Equal(t, int64(777), publishedEvent.Amount)
	assert.Equal(t, "key-pub", publishedEvent.IdempotencyKey)
	assert.False(t, publishedEvent.Timestamp.IsZero())
}

func TestTransfer_Failure_PublishesFailedEvent(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return errors.New("reserve failed")
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	var publishedEvent *domain.TransactionEvent
	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		publishedEvent = event
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         777,
		IdempotencyKey: "key-fail",
	})

	require.Error(t, err)
	require.NotNil(t, publishedEvent)
	assert.Equal(t, domain.EventTransactionFailed, publishedEvent.Type)
	assert.Equal(t, domain.TxStatusFailed, publishedEvent.Status)
	assert.Equal(t, result.ID, publishedEvent.TransactionID)
	assert.Equal(t, "acc-from", publishedEvent.FromAccountID)
	assert.Equal(t, "acc-to", publishedEvent.ToAccountID)
	assert.Equal(t, int64(777), publishedEvent.Amount)
	assert.Equal(t, "key-fail", publishedEvent.IdempotencyKey)
	assert.False(t, publishedEvent.Timestamp.IsZero())
}

func TestTransfer_PublisherNil_NoError(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	svc := NewTransactionService(repo, accountCli, nil)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-nil-pub",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestTransfer_PublisherTimeout_StillSucceeds(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	publisher.PublishTransactionEventFunc = func(event *domain.TransactionEvent) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-slow-pub",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestGetByID_EmptyID(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.GetByIDFunc = func(ctx context.Context, id string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	result, err := svc.GetByID(context.Background(), "")

	require.ErrorIs(t, err, domain.ErrTransactionNotFound)
	assert.Nil(t, result)
}

func TestGetByID_VeryLongID(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	longID := strings.Repeat("x", 10000)

	repo.GetByIDFunc = func(ctx context.Context, id string) (*domain.Transaction, error) {
		if id == longID {
			return &domain.Transaction{
				ID:             longID,
				FromAccountID:  "acc-from",
				ToAccountID:    "acc-to",
				Amount:         100,
				Status:         domain.TxStatusCompleted,
				IdempotencyKey: "key-long",
			}, nil
		}
		return nil, domain.ErrTransactionNotFound
	}

	result, err := svc.GetByID(context.Background(), longID)

	require.NoError(t, err)
	assert.Equal(t, longID, result.ID)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
}

func TestExecuteSaga_AllSucceed(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         1000,
		IdempotencyKey: "key-saga-ok",
	})

	require.NoError(t, err)
	assert.Equal(t, domain.TxStatusCompleted, result.Status)
	assert.Equal(t, int64(1000), result.Amount)
}

func TestExecuteSaga_ReserveFails(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	reserveErr := errors.New("not enough balance")
	accountCli.ReserveFunc = func(ctx context.Context, id string, amount int64) error {
		return reserveErr
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         1000,
		IdempotencyKey: "key-saga-reserve-fail",
	})

	require.Error(t, err)
	assert.Equal(t, reserveErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
}

func TestExecuteSaga_CreditFailsCompensated(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	creditErr := errors.New("credit account unavailable")
	accountCli.CreditFunc = func(ctx context.Context, id string, amount int64) error {
		return creditErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         1000,
		IdempotencyKey: "key-saga-credit-fail",
	})

	require.Error(t, err)
	assert.Equal(t, creditErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.True(t, cancelCalled.Load(), "CancelReservation compensation must be called")
}

func TestExecuteSaga_CommitFailsCompensated(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	publisher := &mocks.MockEventPublisher{}
	svc := NewTransactionService(repo, accountCli, publisher)

	commitErr := errors.New("commit reservation failed")
	accountCli.CommitReservationFunc = func(ctx context.Context, id string, amount int64) error {
		return commitErr
	}

	var cancelCalled atomic.Bool
	accountCli.CancelReservationFunc = func(ctx context.Context, id string, amount int64) error {
		cancelCalled.Store(true)
		return nil
	}

	var debitCalled atomic.Bool
	accountCli.DebitFunc = func(ctx context.Context, id string, amount int64) error {
		debitCalled.Store(true)
		return nil
	}

	repo.CreateFunc = func(ctx context.Context, tx *domain.Transaction) error {
		return nil
	}
	repo.UpdateStatusFunc = func(ctx context.Context, id string, status domain.TransactionStatus) error {
		return nil
	}

	result, err := svc.Transfer(context.Background(), domain.TransferInput{
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         1000,
		IdempotencyKey: "key-saga-commit-fail",
	})

	require.Error(t, err)
	assert.Equal(t, commitErr, err)
	assert.Equal(t, domain.TxStatusFailed, result.Status)
	assert.True(t, cancelCalled.Load(), "CancelReservation compensation must be called")
	assert.True(t, debitCalled.Load(), "Debit compensation must be called")
}
