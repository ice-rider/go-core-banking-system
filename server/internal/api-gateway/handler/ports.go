package handler

import (
	"context"

	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

type AccountClient interface {
	Create(ctx context.Context, ownerName string) (*pb_account.AccountResponse, error)
	GetByID(ctx context.Context, id string) (*pb_account.AccountResponse, error)
	Block(ctx context.Context, id string) (*pb_account.Empty, error)
	Unblock(ctx context.Context, id string) (*pb_account.Empty, error)
	Close(ctx context.Context, id string) (*pb_account.Empty, error)
}

type TransactionClient interface {
	Transfer(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error)
	GetByID(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error)
}
