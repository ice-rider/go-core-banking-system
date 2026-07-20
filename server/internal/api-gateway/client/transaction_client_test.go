package client

import (
	"context"
	"testing"

	"go-core-banking-system/pkg/proto/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTransactionClientWrapper_Transfer(t *testing.T) {
	tests := []struct {
		name           string
		fromAccountID  string
		toAccountID    string
		amount         int64
		idempotencyKey string
		resp           *transaction.TransactionResponse
		mockErr        error
		wantErr        bool
		wantCode       codes.Code
	}{
		{
			name:           "success",
			fromAccountID:  "acc-from-1",
			toAccountID:    "acc-to-2",
			amount:         1000,
			idempotencyKey: "idem-key-123",
			resp: &transaction.TransactionResponse{
				Id:             "txn-001",
				FromAccountId:  "acc-from-1",
				ToAccountId:    "acc-to-2",
				Amount:         1000,
				Status:         "completed",
				IdempotencyKey: "idem-key-123",
			},
		},
		{
			name:           "gRPC not found error",
			fromAccountID:  "nonexistent",
			toAccountID:    "acc-to-2",
			amount:         500,
			idempotencyKey: "idem-key-456",
			mockErr:        status.Error(codes.NotFound, "source account not found"),
			wantErr:        true,
			wantCode:       codes.NotFound,
		},
		{
			name:           "gRPC internal error",
			fromAccountID:  "acc-from-3",
			toAccountID:    "acc-to-4",
			amount:         2000,
			idempotencyKey: "idem-key-789",
			mockErr:        status.Error(codes.Internal, "transfer failed"),
			wantErr:        true,
			wantCode:       codes.Internal,
		},
		{
			name:           "gRPC failed precondition error",
			fromAccountID:  "acc-from-5",
			toAccountID:    "acc-to-6",
			amount:         999999,
			idempotencyKey: "idem-key-abc",
			mockErr:        status.Error(codes.FailedPrecondition, "insufficient funds"),
			wantErr:        true,
			wantCode:       codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTransactionGRPCClient{
				transferFunc: func(ctx context.Context, in *transaction.TransferRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error) {
					require.Equal(t, tt.fromAccountID, in.GetFromAccountId())
					require.Equal(t, tt.toAccountID, in.GetToAccountId())
					require.Equal(t, tt.amount, in.GetAmount())
					require.Equal(t, tt.idempotencyKey, in.GetIdempotencyKey())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewTransactionClientWrapper(mock)
			resp, err := wrapper.Transfer(context.Background(), tt.fromAccountID, tt.toAccountID, tt.amount, tt.idempotencyKey)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "txn-001", resp.GetId())
				assert.Equal(t, "acc-from-1", resp.GetFromAccountId())
				assert.Equal(t, "acc-to-2", resp.GetToAccountId())
				assert.Equal(t, int64(1000), resp.GetAmount())
				assert.Equal(t, "completed", resp.GetStatus())
				assert.Equal(t, "idem-key-123", resp.GetIdempotencyKey())
			}
		})
	}
}

func TestTransactionClientWrapper_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		resp     *transaction.TransactionResponse
		mockErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "success",
			id:   "txn-001",
			resp: &transaction.TransactionResponse{
				Id:             "txn-001",
				FromAccountId:  "acc-from-1",
				ToAccountId:    "acc-to-2",
				Amount:         1000,
				Status:         "completed",
				IdempotencyKey: "idem-key-123",
			},
		},
		{
			name:     "gRPC not found error",
			id:       "nonexistent",
			mockErr:  status.Error(codes.NotFound, "transaction not found"),
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name:     "gRPC internal error",
			id:       "txn-error",
			mockErr:  status.Error(codes.Internal, "internal failure"),
			wantErr:  true,
			wantCode: codes.Internal,
		},
		{
			name:     "gRPC permission denied error",
			id:       "txn-unauthorized",
			mockErr:  status.Error(codes.PermissionDenied, "access denied"),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTransactionGRPCClient{
				getByIDFunc: func(ctx context.Context, in *transaction.GetTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionResponse, error) {
					require.Equal(t, tt.id, in.GetId())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewTransactionClientWrapper(mock)
			resp, err := wrapper.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "txn-001", resp.GetId())
				assert.Equal(t, "acc-from-1", resp.GetFromAccountId())
				assert.Equal(t, "acc-to-2", resp.GetToAccountId())
				assert.Equal(t, int64(1000), resp.GetAmount())
				assert.Equal(t, "completed", resp.GetStatus())
				assert.Equal(t, "idem-key-123", resp.GetIdempotencyKey())
			}
		})
	}
}
