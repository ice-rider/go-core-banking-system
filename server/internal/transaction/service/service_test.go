package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/transaction/domain"
	"go-core-banking-system/internal/transaction/mocks"
)

func TestTransfer_ValidTransfer(t *testing.T) {
	repo := &mocks.MockTransactionRepository{}
	accountCli := &mocks.MockAccountClient{}
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
		return errors.New("duplicate key violates unique constraint unique_violation")
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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

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
	svc := NewTransactionService(repo, accountCli)

	repo.GetByIDFunc = func(ctx context.Context, id string) (*domain.Transaction, error) {
		return nil, domain.ErrTransactionNotFound
	}

	result, err := svc.GetByID(context.Background(), "nonexistent-id")

	require.ErrorIs(t, err, domain.ErrTransactionNotFound)
	assert.Nil(t, result)
}
