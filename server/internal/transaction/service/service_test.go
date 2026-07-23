package service

import (
	"context"
	"errors"
	"testing"

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
