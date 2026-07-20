package client

import (
	"context"

	"go-core-banking-system/pkg/proto/transaction"
)

type TransactionClientWrapper struct {
	client transaction.TransactionServiceClient
}

func NewTransactionClientWrapper(client transaction.TransactionServiceClient) *TransactionClientWrapper {
	return &TransactionClientWrapper{client: client}
}

func (w *TransactionClientWrapper) Transfer(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*transaction.TransactionResponse, error) {
	return w.client.Transfer(ctx, &transaction.TransferRequest{
		FromAccountId:  fromAccountID,
		ToAccountId:    toAccountID,
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
	})
}

func (w *TransactionClientWrapper) GetByID(ctx context.Context, id string) (*transaction.TransactionResponse, error) {
	return w.client.GetByID(ctx, &transaction.GetTransactionRequest{Id: id})
}
