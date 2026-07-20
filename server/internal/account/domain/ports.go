package domain

import "context"

type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	UpdateStatus(ctx context.Context, id string, status Status) error
	Reserve(ctx context.Context, id string, amount int64) error
	Credit(ctx context.Context, id string, amount int64) error
	Debit(ctx context.Context, id string, amount int64) error
	CommitReservation(ctx context.Context, id string, amount int64) error
	CancelReservation(ctx context.Context, id string, amount int64) error
	BlockAtomically(ctx context.Context, id string) error
	UnblockAtomically(ctx context.Context, id string) error
	CloseAtomically(ctx context.Context, id string) error
}

type Service interface {
	Create(ctx context.Context, input CreateAccountInput) (*Account, error)
	GetByID(ctx context.Context, id string) (*Account, error)
	Block(ctx context.Context, id string) error
	Unblock(ctx context.Context, id string) error
	Close(ctx context.Context, id string) error
	Reserve(ctx context.Context, id string, amount int64) error
	Credit(ctx context.Context, id string, amount int64) error
	Debit(ctx context.Context, id string, amount int64) error
	CommitReservation(ctx context.Context, id string, amount int64) error
	CancelReservation(ctx context.Context, id string, amount int64) error
}

type CreateAccountInput struct {
	OwnerName string
}
