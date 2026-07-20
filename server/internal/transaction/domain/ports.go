package domain

import "context"

type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	UpdateStatus(ctx context.Context, id string, status TransactionStatus) error
	GetByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
}

type Service interface {
	Transfer(ctx context.Context, input TransferInput) (*Transaction, error)
	GetByID(ctx context.Context, id string) (*Transaction, error)
}

type TransferInput struct {
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	IdempotencyKey string
}
