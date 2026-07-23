package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/transaction/domain"
	"go-core-banking-system/pkg/proto/transaction"
)

type mockTransactionService struct {
	TransferFunc func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error)
	GetByIDFunc  func(ctx context.Context, id string) (*domain.Transaction, error)
}

func (m *mockTransactionService) Transfer(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
	if m.TransferFunc != nil {
		return m.TransferFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockTransactionService) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func testTime() time.Time {
	return time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
}

func testTransaction() *domain.Transaction {
	return &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "idem-1",
		CreatedAt:      testTime(),
		UpdatedAt:      testTime(),
	}
}

func TestTransfer_Success(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			assert.Equal(t, "acc-from", input.FromAccountID)
			assert.Equal(t, "acc-to", input.ToAccountID)
			assert.Equal(t, int64(500), input.Amount)
			assert.Equal(t, "idem-1", input.IdempotencyKey)
			return testTransaction(), nil
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "idem-1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "tx-1", resp.Id)
	assert.Equal(t, "acc-from", resp.FromAccountId)
	assert.Equal(t, "acc-to", resp.ToAccountId)
	assert.Equal(t, int64(500), resp.Amount)
	assert.Equal(t, "COMPLETED", resp.Status)
	assert.Equal(t, "idem-1", resp.IdempotencyKey)
	assert.NotEmpty(t, resp.CreatedAt)
	assert.NotEmpty(t, resp.UpdatedAt)
}

func TestTransfer_SameAccount(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrSameAccount
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-1",
		ToAccountId:    "acc-1",
		Amount:         500,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrSameAccount)
}

func TestTransfer_InvalidAmount(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrInvalidAmount
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         -100,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrInsufficientFunds
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         10000,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestTransfer_AccountNotFound(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrAccountNotFound
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "nonexistent",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestTransfer_AccountBlocked(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrAccountBlocked
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-blocked",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrAccountBlocked)
}

func TestTransfer_IdempotencyKeyUsed(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrIdempotencyKeyUsed
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "duplicate-key",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrIdempotencyKeyUsed)
}

func TestTransfer_IdempotencyKeyRequired(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, domain.ErrIdempotencyKeyRequired
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrIdempotencyKeyRequired)
}

func TestTransfer_ServiceError(t *testing.T) {
	svc := &mockTransactionService{
		TransferFunc: func(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
			return nil, errors.New("unexpected error")
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.Transfer(context.Background(), &transaction.TransferRequest{
		FromAccountId:  "acc-from",
		ToAccountId:    "acc-to",
		Amount:         500,
		IdempotencyKey: "idem-1",
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "unexpected error")
}

func TestGetByID_Success(t *testing.T) {
	svc := &mockTransactionService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Transaction, error) {
			assert.Equal(t, "tx-1", id)
			return testTransaction(), nil
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &transaction.GetTransactionRequest{Id: "tx-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "tx-1", resp.Id)
	assert.Equal(t, "acc-from", resp.FromAccountId)
	assert.Equal(t, "acc-to", resp.ToAccountId)
	assert.Equal(t, int64(500), resp.Amount)
	assert.Equal(t, "COMPLETED", resp.Status)
	assert.Equal(t, "idem-1", resp.IdempotencyKey)
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockTransactionService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return nil, domain.ErrTransactionNotFound
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &transaction.GetTransactionRequest{Id: "nonexistent"})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestGetByID_RepoError(t *testing.T) {
	svc := &mockTransactionService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Transaction, error) {
			return nil, errors.New("db connection refused")
		},
	}
	h := NewTransactionGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &transaction.GetTransactionRequest{Id: "tx-1"})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "db connection refused")
}

func TestNewTransactionGRPCHandler(t *testing.T) {
	svc := &mockTransactionService{}
	h := NewTransactionGRPCHandler(svc)

	require.NotNil(t, h)
	assert.Equal(t, svc, h.svc)
}
