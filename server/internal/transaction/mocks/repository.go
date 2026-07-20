package mocks

import (
	"context"

	"go-core-banking-system/internal/transaction/domain"
)

type MockTransactionRepository struct {
	CreateFunc              func(ctx context.Context, tx *domain.Transaction) error
	GetByIDFunc             func(ctx context.Context, id string) (*domain.Transaction, error)
	UpdateStatusFunc        func(ctx context.Context, id string, status domain.TransactionStatus) error
	GetByIdempotencyKeyFunc func(ctx context.Context, key string) (*domain.Transaction, error)
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tx)
	}
	return nil
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrTransactionNotFound
}

func (m *MockTransactionRepository) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockTransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	if m.GetByIdempotencyKeyFunc != nil {
		return m.GetByIdempotencyKeyFunc(ctx, key)
	}
	return nil, domain.ErrTransactionNotFound
}
