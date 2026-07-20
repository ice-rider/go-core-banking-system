package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/api-gateway/validator"
	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type accountResponse struct {
	ID        string `json:"id"`
	OwnerName string `json:"owner_name"`
	Balance   int64  `json:"balance"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type transferResponse struct {
	ID             string `json:"id"`
	FromAccountID  string `json:"from_account_id"`
	ToAccountID    string `json:"to_account_id"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

type messageResponse struct {
	Message string `json:"message"`
}

// ==================== Mock Account Client ====================

type mockAccountClient struct {
	CreateFunc  func(ctx context.Context, ownerName string) (*pb_account.AccountResponse, error)
	GetByIDFunc func(ctx context.Context, id string) (*pb_account.AccountResponse, error)
	BlockFunc   func(ctx context.Context, id string) (*pb_account.Empty, error)
	UnblockFunc func(ctx context.Context, id string) (*pb_account.Empty, error)
	CloseFunc   func(ctx context.Context, id string) (*pb_account.Empty, error)
}

func (m *mockAccountClient) Create(ctx context.Context, ownerName string) (*pb_account.AccountResponse, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, ownerName)
	}
	return nil, nil
}

func (m *mockAccountClient) GetByID(ctx context.Context, id string) (*pb_account.AccountResponse, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAccountClient) Block(ctx context.Context, id string) (*pb_account.Empty, error) {
	if m.BlockFunc != nil {
		return m.BlockFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAccountClient) Unblock(ctx context.Context, id string) (*pb_account.Empty, error) {
	if m.UnblockFunc != nil {
		return m.UnblockFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAccountClient) Close(ctx context.Context, id string) (*pb_account.Empty, error) {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id)
	}
	return nil, nil
}

// ==================== Mock Transaction Client ====================

type mockTransactionClient struct {
	TransferFunc func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error)
	GetByIDFunc  func(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error)
}

func (m *mockTransactionClient) Transfer(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
	if m.TransferFunc != nil {
		return m.TransferFunc(ctx, fromAccountID, toAccountID, amount, idempotencyKey)
	}
	return nil, nil
}

func (m *mockTransactionClient) GetByID(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

// ==================== Test Helpers ====================

func newHandler(ac AccountClient, tc TransactionClient) *Handler {
	return NewHandler(ac, tc, validator.New())
}

func newTestContext(method, path, body string, params ...string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i < len(params)-1; i += 2 {
		c.AddParam(params[i], params[i+1])
	}
	return c, w
}

func mustUnmarshalAccount(t *testing.T, data []byte) accountResponse {
	t.Helper()
	var resp accountResponse
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

func mustUnmarshalTransfer(t *testing.T, data []byte) transferResponse {
	t.Helper()
	var resp transferResponse
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

func mustUnmarshalError(t *testing.T, data []byte) errorResponse {
	t.Helper()
	var resp errorResponse
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

func mustUnmarshalMessage(t *testing.T, data []byte) messageResponse {
	t.Helper()
	var resp messageResponse
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

// ==================== POST /accounts ====================

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func() *mockAccountClient
		wantStatusCode int
		wantAccount    func(t *testing.T, resp accountResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name: "success",
			body: `{"owner_name": "Alice"}`,
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CreateFunc: func(ctx context.Context, ownerName string) (*pb_account.AccountResponse, error) {
						assert.Equal(t, "Alice", ownerName)
						return &pb_account.AccountResponse{
							Id: "acc-1", OwnerName: "Alice", Balance: 0, Status: "active",
							CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
						}, nil
					},
				}
			},
			wantStatusCode: http.StatusCreated,
			wantAccount: func(t *testing.T, resp accountResponse) {
				assert.Equal(t, "acc-1", resp.ID)
				assert.Equal(t, "Alice", resp.OwnerName)
				assert.Equal(t, int64(0), resp.Balance)
				assert.Equal(t, "active", resp.Status)
			},
		},
		{
			name:           "bad JSON body",
			body:           `{invalid`,
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:           "empty owner name",
			body:           `{"owner_name": ""}`,
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
				assert.NotEmpty(t, resp.Error)
			},
		},
		{
			name:           "missing owner name field",
			body:           `{}`,
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name: "client error returns 500",
			body: `{"owner_name": "Bob"}`,
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CreateFunc: func(ctx context.Context, ownerName string) (*pb_account.AccountResponse, error) {
						return nil, errors.New("service unavailable")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.mockSetup(), &mockTransactionClient{})

			c, w := newTestContext("POST", "/accounts", tt.body)
			h.CreateAccount(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantAccount != nil {
				tt.wantAccount(t, mustUnmarshalAccount(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== GET /accounts/:id ====================

func TestGetAccount(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockSetup      func() *mockAccountClient
		wantStatusCode int
		wantAccount    func(t *testing.T, resp accountResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name:      "success",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_account.AccountResponse, error) {
						assert.Equal(t, "acc-1", id)
						return &pb_account.AccountResponse{
							Id: "acc-1", OwnerName: "Alice", Balance: 500, Status: "active",
							CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
						}, nil
					},
				}
			},
			wantStatusCode: http.StatusOK,
			wantAccount: func(t *testing.T, resp accountResponse) {
				assert.Equal(t, "acc-1", resp.ID)
				assert.Equal(t, "Alice", resp.OwnerName)
				assert.Equal(t, int64(500), resp.Balance)
			},
		},
		{
			name:      "account not found",
			pathParam: "nonexistent",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_account.AccountResponse, error) {
						return nil, errors.New("account not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
				assert.Contains(t, resp.Error, "account not found")
			},
		},
		{
			name:           "missing id param",
			pathParam:      "",
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:      "client error returns 500",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_account.AccountResponse, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.mockSetup(), &mockTransactionClient{})

			c, w := newTestContext("GET", "/accounts/"+tt.pathParam, "", "id", tt.pathParam)
			h.GetAccount(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantAccount != nil {
				tt.wantAccount(t, mustUnmarshalAccount(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== POST /accounts/:id/block ====================

func TestBlockAccount(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockSetup      func() *mockAccountClient
		wantStatusCode int
		wantMessage    func(t *testing.T, resp messageResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name:      "success",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					BlockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						assert.Equal(t, "acc-1", id)
						return &pb_account.Empty{}, nil
					},
				}
			},
			wantStatusCode: http.StatusOK,
			wantMessage: func(t *testing.T, resp messageResponse) {
				assert.Equal(t, "account blocked", resp.Message)
			},
		},
		{
			name:      "account not found",
			pathParam: "nonexistent",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					BlockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
				assert.Contains(t, resp.Error, "account not found")
			},
		},
		{
			name:      "account already blocked",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					BlockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account is blocked")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "account is blocked")
			},
		},
		{
			name:           "missing id param",
			pathParam:      "",
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:      "client error returns 500",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					BlockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.mockSetup(), &mockTransactionClient{})

			c, w := newTestContext("POST", "/accounts/"+tt.pathParam+"/block", "", "id", tt.pathParam)
			h.BlockAccount(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantMessage != nil {
				tt.wantMessage(t, mustUnmarshalMessage(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== POST /accounts/:id/unblock ====================

func TestUnblockAccount(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockSetup      func() *mockAccountClient
		wantStatusCode int
		wantMessage    func(t *testing.T, resp messageResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name:      "success",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					UnblockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						assert.Equal(t, "acc-1", id)
						return &pb_account.Empty{}, nil
					},
				}
			},
			wantStatusCode: http.StatusOK,
			wantMessage: func(t *testing.T, resp messageResponse) {
				assert.Equal(t, "account unblocked", resp.Message)
			},
		},
		{
			name:      "account not found",
			pathParam: "nonexistent",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					UnblockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
				assert.Contains(t, resp.Error, "account not found")
			},
		},
		{
			name:      "account not active",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					UnblockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account is not active")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "not active")
			},
		},
		{
			name:           "missing id param",
			pathParam:      "",
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:      "client error returns 500",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					UnblockFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.mockSetup(), &mockTransactionClient{})

			c, w := newTestContext("POST", "/accounts/"+tt.pathParam+"/unblock", "", "id", tt.pathParam)
			h.UnblockAccount(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantMessage != nil {
				tt.wantMessage(t, mustUnmarshalMessage(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== POST /accounts/:id/close ====================

func TestCloseAccount(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockSetup      func() *mockAccountClient
		wantStatusCode int
		wantMessage    func(t *testing.T, resp messageResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name:      "success",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CloseFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						assert.Equal(t, "acc-1", id)
						return &pb_account.Empty{}, nil
					},
				}
			},
			wantStatusCode: http.StatusOK,
			wantMessage: func(t *testing.T, resp messageResponse) {
				assert.Equal(t, "account closed", resp.Message)
			},
		},
		{
			name:      "account not found",
			pathParam: "nonexistent",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CloseFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
				assert.Contains(t, resp.Error, "account not found")
			},
		},
		{
			name:      "balance not zero",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CloseFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("balance must be zero to close account")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "balance must be zero")
			},
		},
		{
			name:      "account already closed",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CloseFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("account is closed")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "account is closed")
			},
		},
		{
			name:           "missing id param",
			pathParam:      "",
			mockSetup:      func() *mockAccountClient { return &mockAccountClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:      "client error returns 500",
			pathParam: "acc-1",
			mockSetup: func() *mockAccountClient {
				return &mockAccountClient{
					CloseFunc: func(ctx context.Context, id string) (*pb_account.Empty, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.mockSetup(), &mockTransactionClient{})

			c, w := newTestContext("POST", "/accounts/"+tt.pathParam+"/close", "", "id", tt.pathParam)
			h.CloseAccount(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantMessage != nil {
				tt.wantMessage(t, mustUnmarshalMessage(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== POST /transfers ====================

func TestCreateTransfer(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockSetup      func() *mockTransactionClient
		wantStatusCode int
		wantTransfer   func(t *testing.T, resp transferResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name: "success",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", fromAccountID)
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", toAccountID)
						assert.Equal(t, int64(100), amount)
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440003", idempotencyKey)
						return &pb_transaction.TransactionResponse{
							Id: "tx-1", FromAccountId: "550e8400-e29b-41d4-a716-446655440001",
							ToAccountId: "550e8400-e29b-41d4-a716-446655440002",
							Amount:      100, Status: "completed", IdempotencyKey: "550e8400-e29b-41d4-a716-446655440003",
							CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
						}, nil
					},
				}
			},
			wantStatusCode: http.StatusCreated,
			wantTransfer: func(t *testing.T, resp transferResponse) {
				assert.Equal(t, "tx-1", resp.ID)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", resp.FromAccountID)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440002", resp.ToAccountID)
				assert.Equal(t, int64(100), resp.Amount)
				assert.Equal(t, "completed", resp.Status)
			},
		},
		{
			name:           "bad JSON body",
			body:           `{invalid`,
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:           "invalid amount zero",
			body:           `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":0,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:           "invalid amount negative",
			body:           `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":-50,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:           "empty idempotency key",
			body:           `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":""}`,
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:           "missing required fields",
			body:           `{}`,
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name: "same account",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440001","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("cannot transfer to the same account")
					},
				}
			},
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
				assert.Contains(t, resp.Error, "same account")
			},
		},
		{
			name: "account not found",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("account not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
			},
		},
		{
			name: "insufficient funds",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":99999,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("insufficient funds")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "insufficient funds")
			},
		},
		{
			name: "account blocked",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("account is blocked")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "account is blocked")
			},
		},
		{
			name: "idempotency key already used",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("idempotency key already used")
					},
				}
			},
			wantStatusCode: http.StatusConflict,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "CONFLICT", resp.Code)
				assert.Contains(t, resp.Error, "idempotency key already used")
			},
		},
		{
			name: "client error returns 500",
			body: `{"from_account_id":"550e8400-e29b-41d4-a716-446655440001","to_account_id":"550e8400-e29b-41d4-a716-446655440002","amount":100,"idempotency_key":"550e8400-e29b-41d4-a716-446655440003"}`,
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					TransferFunc: func(ctx context.Context, fromAccountID, toAccountID string, amount int64, idempotencyKey string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(&mockAccountClient{}, tt.mockSetup())

			c, w := newTestContext("POST", "/transfers", tt.body)
			h.CreateTransfer(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantTransfer != nil {
				tt.wantTransfer(t, mustUnmarshalTransfer(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}

// ==================== GET /transfers/:id ====================

func TestGetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		pathParam      string
		mockSetup      func() *mockTransactionClient
		wantStatusCode int
		wantTransfer   func(t *testing.T, resp transferResponse)
		wantError      func(t *testing.T, resp errorResponse)
	}{
		{
			name:      "success",
			pathParam: "tx-1",
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error) {
						assert.Equal(t, "tx-1", id)
						return &pb_transaction.TransactionResponse{
							Id: "tx-1", FromAccountId: "acc-1", ToAccountId: "acc-2",
							Amount: 100, Status: "completed", IdempotencyKey: "key-1",
							CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-01T00:00:00Z",
						}, nil
					},
				}
			},
			wantStatusCode: http.StatusOK,
			wantTransfer: func(t *testing.T, resp transferResponse) {
				assert.Equal(t, "tx-1", resp.ID)
				assert.Equal(t, "acc-1", resp.FromAccountID)
				assert.Equal(t, "acc-2", resp.ToAccountID)
				assert.Equal(t, int64(100), resp.Amount)
				assert.Equal(t, "completed", resp.Status)
			},
		},
		{
			name:      "transaction not found",
			pathParam: "nonexistent",
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("transaction not found")
					},
				}
			},
			wantStatusCode: http.StatusNotFound,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "NOT_FOUND", resp.Code)
				assert.Contains(t, resp.Error, "transaction not found")
			},
		},
		{
			name:           "missing id param",
			pathParam:      "",
			mockSetup:      func() *mockTransactionClient { return &mockTransactionClient{} },
			wantStatusCode: http.StatusBadRequest,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "VALIDATION_ERROR", resp.Code)
			},
		},
		{
			name:      "client error returns 500",
			pathParam: "tx-1",
			mockSetup: func() *mockTransactionClient {
				return &mockTransactionClient{
					GetByIDFunc: func(ctx context.Context, id string) (*pb_transaction.TransactionResponse, error) {
						return nil, errors.New("connection refused")
					},
				}
			},
			wantStatusCode: http.StatusInternalServerError,
			wantError: func(t *testing.T, resp errorResponse) {
				assert.Equal(t, "INTERNAL_ERROR", resp.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(&mockAccountClient{}, tt.mockSetup())

			c, w := newTestContext("GET", "/transfers/"+tt.pathParam, "", "id", tt.pathParam)
			h.GetTransaction(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantTransfer != nil {
				tt.wantTransfer(t, mustUnmarshalTransfer(t, w.Body.Bytes()))
			}
			if tt.wantError != nil {
				tt.wantError(t, mustUnmarshalError(t, w.Body.Bytes()))
			}
		})
	}
}
