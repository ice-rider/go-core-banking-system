package client

import (
	"context"

	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

type AccountClient interface {
	Create(ctx context.Context, req *pb_account.CreateAccountRequest) (*pb_account.AccountResponse, error)
	GetByID(ctx context.Context, req *pb_account.GetAccountRequest) (*pb_account.AccountResponse, error)
	Block(ctx context.Context, req *pb_account.BlockAccountRequest) (*pb_account.Empty, error)
	Unblock(ctx context.Context, req *pb_account.UnblockAccountRequest) (*pb_account.Empty, error)
	Close(ctx context.Context, req *pb_account.CloseAccountRequest) (*pb_account.Empty, error)
}

type TransactionClient interface {
	Transfer(ctx context.Context, req *pb_transaction.TransferRequest) (*pb_transaction.TransactionResponse, error)
	GetByID(ctx context.Context, req *pb_transaction.GetTransactionRequest) (*pb_transaction.TransactionResponse, error)
}
