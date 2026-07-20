package client

import (
	"context"

	"go-core-banking-system/pkg/proto/account"
)

type AccountClientWrapper struct {
	client account.AccountServiceClient
}

func NewAccountClientWrapper(client account.AccountServiceClient) *AccountClientWrapper {
	return &AccountClientWrapper{client: client}
}

func (w *AccountClientWrapper) Create(ctx context.Context, ownerName string) (*account.AccountResponse, error) {
	return w.client.Create(ctx, &account.CreateAccountRequest{OwnerName: ownerName})
}

func (w *AccountClientWrapper) GetByID(ctx context.Context, id string) (*account.AccountResponse, error) {
	return w.client.GetByID(ctx, &account.GetAccountRequest{Id: id})
}

func (w *AccountClientWrapper) Block(ctx context.Context, id string) (*account.Empty, error) {
	return w.client.Block(ctx, &account.BlockAccountRequest{Id: id})
}

func (w *AccountClientWrapper) Unblock(ctx context.Context, id string) (*account.Empty, error) {
	return w.client.Unblock(ctx, &account.UnblockAccountRequest{Id: id})
}

func (w *AccountClientWrapper) Close(ctx context.Context, id string) (*account.Empty, error) {
	return w.client.Close(ctx, &account.CloseAccountRequest{Id: id})
}
