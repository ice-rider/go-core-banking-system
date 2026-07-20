package client

import (
	"context"

	"go-core-banking-system/pkg/proto/account"
	"go-core-banking-system/pkg/proto/transaction"

	"google.golang.org/grpc"
)

type mockAccountGRPCClient struct {
	account.AccountServiceClient
	createFunc            func(ctx context.Context, in *account.CreateAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error)
	getByIDFunc           func(ctx context.Context, in *account.GetAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error)
	blockFunc             func(ctx context.Context, in *account.BlockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error)
	unblockFunc           func(ctx context.Context, in *account.UnblockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error)
	closeFunc             func(ctx context.Context, in *account.CloseAccountRequest, opts ...grpc.CallOption) (*account.Empty, error)
	reserveFunc           func(ctx context.Context, in *account.ReserveRequest, opts ...grpc.CallOption) (*account.Empty, error)
	creditFunc            func(ctx context.Context, in *account.CreditRequest, opts ...grpc.CallOption) (*account.Empty, error)
	debitFunc             func(ctx context.Context, in *account.DebitRequest, opts ...grpc.CallOption) (*account.Empty, error)
	commitReservationFunc func(ctx context.Context, in *account.CommitReservationRequest, opts ...grpc.CallOption) (*account.Empty, error)
	cancelReservationFunc func(ctx context.Context, in *account.CancelReservationRequest, opts ...grpc.CallOption) (*account.Empty, error)
}

func (m *mockAccountGRPCClient) Create(ctx context.Context, in *account.CreateAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, in, opts...)
	}
	panic("createFunc not set")
}

func (m *mockAccountGRPCClient) GetByID(ctx context.Context, in *account.GetAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, in, opts...)
	}
	panic("getByIDFunc not set")
}

func (m *mockAccountGRPCClient) Block(ctx context.Context, in *account.BlockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.blockFunc != nil {
		return m.blockFunc(ctx, in, opts...)
	}
	panic("blockFunc not set")
}

func (m *mockAccountGRPCClient) Unblock(ctx context.Context, in *account.UnblockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.unblockFunc != nil {
		return m.unblockFunc(ctx, in, opts...)
	}
	panic("unblockFunc not set")
}

func (m *mockAccountGRPCClient) Close(ctx context.Context, in *account.CloseAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.closeFunc != nil {
		return m.closeFunc(ctx, in, opts...)
	}
	panic("closeFunc not set")
}

func (m *mockAccountGRPCClient) Reserve(ctx context.Context, in *account.ReserveRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.reserveFunc != nil {
		return m.reserveFunc(ctx, in, opts...)
	}
	panic("reserveFunc not set")
}

func (m *mockAccountGRPCClient) Credit(ctx context.Context, in *account.CreditRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.creditFunc != nil {
		return m.creditFunc(ctx, in, opts...)
	}
	panic("creditFunc not set")
}

func (m *mockAccountGRPCClient) Debit(ctx context.Context, in *account.DebitRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.debitFunc != nil {
		return m.debitFunc(ctx, in, opts...)
	}
	panic("debitFunc not set")
}

func (m *mockAccountGRPCClient) CommitReservation(ctx context.Context, in *account.CommitReservationRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.commitReservationFunc != nil {
		return m.commitReservationFunc(ctx, in, opts...)
	}
	panic("commitReservationFunc not set")
}

func (m *mockAccountGRPCClient) CancelReservation(ctx context.Context, in *account.CancelReservationRequest, opts ...grpc.CallOption) (*account.Empty, error) {
	if m.cancelReservationFunc != nil {
		return m.cancelReservationFunc(ctx, in, opts...)
	}
	panic("cancelReservationFunc not set")
}

type mockTransactionGRPCClient struct {
	transaction.TransactionServiceClient
	transferFunc func(ctx context.Context, in *transaction.TransferRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error)
	getByIDFunc  func(ctx context.Context, in *transaction.GetTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error)
}

func (m *mockTransactionGRPCClient) Transfer(ctx context.Context, in *transaction.TransferRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error) {
	if m.transferFunc != nil {
		return m.transferFunc(ctx, in, opts...)
	}
	panic("transferFunc not set")
}

func (m *mockTransactionGRPCClient) GetByID(ctx context.Context, in *transaction.GetTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, in, opts...)
	}
	panic("getByIDFunc not set")
}
