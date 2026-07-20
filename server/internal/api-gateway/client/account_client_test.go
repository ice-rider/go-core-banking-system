package client

import (
	"context"
	"testing"

	"go-core-banking-system/pkg/proto/account"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccountClientWrapper_Create(t *testing.T) {
	tests := []struct {
		name      string
		ownerName string
		resp      *account.AccountResponse
		mockErr   error
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name:      "success",
			ownerName: "John Doe",
			resp: &account.AccountResponse{
				Id:        "acc-123",
				OwnerName: "John Doe",
				Balance:   0,
				Status:    "active",
			},
		},
		{
			name:      "gRPC not found error",
			ownerName: "Nobody",
			mockErr:   status.Error(codes.NotFound, "account not found"),
			wantErr:   true,
			wantCode:  codes.NotFound,
		},
		{
			name:      "gRPC internal error",
			ownerName: "Anyone",
			mockErr:   status.Error(codes.Internal, "internal failure"),
			wantErr:   true,
			wantCode:  codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountGRPCClient{
				createFunc: func(ctx context.Context, in *account.CreateAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error) {
					require.Equal(t, tt.ownerName, in.GetOwnerName())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewAccountClientWrapper(mock)
			resp, err := wrapper.Create(context.Background(), tt.ownerName)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "acc-123", resp.GetId())
				assert.Equal(t, "John Doe", resp.GetOwnerName())
			}
		})
	}
}

func TestAccountClientWrapper_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		resp     *account.AccountResponse
		mockErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "success",
			id:   "acc-456",
			resp: &account.AccountResponse{
				Id:        "acc-456",
				OwnerName: "Jane Smith",
				Balance:   5000,
				Status:    "active",
			},
		},
		{
			name:     "gRPC not found error",
			id:       "nonexistent",
			mockErr:  status.Error(codes.NotFound, "account not found"),
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name:     "gRPC permission denied error",
			id:       "acc-789",
			mockErr:  status.Error(codes.PermissionDenied, "access denied"),
			wantErr:  true,
			wantCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountGRPCClient{
				getByIDFunc: func(ctx context.Context, in *account.GetAccountRequest, opts ...grpc.CallOption) (*account.AccountResponse, error) {
					require.Equal(t, tt.id, in.GetId())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewAccountClientWrapper(mock)
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
				assert.Equal(t, "acc-456", resp.GetId())
				assert.Equal(t, "Jane Smith", resp.GetOwnerName())
				assert.Equal(t, int64(5000), resp.GetBalance())
			}
		})
	}
}

func TestAccountClientWrapper_Block(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		resp     *account.Empty
		mockErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "success",
			id:   "acc-100",
			resp: &account.Empty{},
		},
		{
			name:     "gRPC not found error",
			id:       "nonexistent",
			mockErr:  status.Error(codes.NotFound, "account not found"),
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name:     "gRPC internal error",
			id:       "acc-200",
			mockErr:  status.Error(codes.Internal, "internal failure"),
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountGRPCClient{
				blockFunc: func(ctx context.Context, in *account.BlockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
					require.Equal(t, tt.id, in.GetId())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewAccountClientWrapper(mock)
			resp, err := wrapper.Block(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

func TestAccountClientWrapper_Unblock(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		resp     *account.Empty
		mockErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "success",
			id:   "acc-100",
			resp: &account.Empty{},
		},
		{
			name:     "gRPC not found error",
			id:       "nonexistent",
			mockErr:  status.Error(codes.NotFound, "account not found"),
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name:     "gRPC failed precondition error",
			id:       "acc-300",
			mockErr:  status.Error(codes.FailedPrecondition, "account is not blocked"),
			wantErr:  true,
			wantCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountGRPCClient{
				unblockFunc: func(ctx context.Context, in *account.UnblockAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
					require.Equal(t, tt.id, in.GetId())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewAccountClientWrapper(mock)
			resp, err := wrapper.Unblock(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}

func TestAccountClientWrapper_Close(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		resp     *account.Empty
		mockErr  error
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "success",
			id:   "acc-500",
			resp: &account.Empty{},
		},
		{
			name:     "gRPC not found error",
			id:       "nonexistent",
			mockErr:  status.Error(codes.NotFound, "account not found"),
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name:     "gRPC failed precondition error",
			id:       "acc-600",
			mockErr:  status.Error(codes.FailedPrecondition, "account has non-zero balance"),
			wantErr:  true,
			wantCode: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAccountGRPCClient{
				closeFunc: func(ctx context.Context, in *account.CloseAccountRequest, opts ...grpc.CallOption) (*account.Empty, error) {
					require.Equal(t, tt.id, in.GetId())
					return tt.resp, tt.mockErr
				},
			}

			wrapper := NewAccountClientWrapper(mock)
			resp, err := wrapper.Close(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
			}
		})
	}
}
