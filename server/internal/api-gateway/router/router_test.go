package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"go-core-banking-system/internal/api-gateway/handler"
	"go-core-banking-system/internal/api-gateway/validator"
	pb_account "go-core-banking-system/pkg/proto/account"
	pb_transaction "go-core-banking-system/pkg/proto/transaction"
)

type mockAccountClient struct{}

func (m *mockAccountClient) Create(_ context.Context, _ string) (*pb_account.AccountResponse, error) {
	return &pb_account.AccountResponse{Id: "1", OwnerName: "Test"}, nil
}
func (m *mockAccountClient) GetByID(_ context.Context, _ string) (*pb_account.AccountResponse, error) {
	return &pb_account.AccountResponse{Id: "1"}, nil
}
func (m *mockAccountClient) Block(_ context.Context, _ string) (*pb_account.Empty, error) {
	return &pb_account.Empty{}, nil
}
func (m *mockAccountClient) Unblock(_ context.Context, _ string) (*pb_account.Empty, error) {
	return &pb_account.Empty{}, nil
}
func (m *mockAccountClient) Close(_ context.Context, _ string) (*pb_account.Empty, error) {
	return &pb_account.Empty{}, nil
}

type mockTransactionClient struct{}

func (m *mockTransactionClient) Transfer(_ context.Context, _, _ string, _ int64, _ string) (*pb_transaction.TransactionResponse, error) {
	return &pb_transaction.TransactionResponse{Id: "1"}, nil
}
func (m *mockTransactionClient) GetByID(_ context.Context, _ string) (*pb_transaction.TransactionResponse, error) {
	return &pb_transaction.TransactionResponse{Id: "1"}, nil
}

func newTestRouter() *gin.Engine {
	h := handler.NewHandler(&mockAccountClient{}, &mockTransactionClient{}, validator.New())
	return NewRouter(h)
}

func TestNewRouter_RoutesExist(t *testing.T) {
	r := newTestRouter()

	routes := r.Routes()
	routeMap := make(map[string]bool)
	for _, route := range routes {
		routeMap[route.Method+":"+route.Path] = true
	}

	assert.True(t, routeMap["POST:/accounts"])
	assert.True(t, routeMap["GET:/accounts/:id"])
	assert.True(t, routeMap["POST:/accounts/:id/block"])
	assert.True(t, routeMap["POST:/accounts/:id/unblock"])
	assert.True(t, routeMap["POST:/accounts/:id/close"])
	assert.True(t, routeMap["POST:/transfers"])
	assert.True(t, routeMap["GET:/transfers/:id"])
	assert.True(t, routeMap["GET:/healthz"])
	assert.True(t, routeMap["GET:/readyz"])
	assert.True(t, routeMap["GET:/metrics"])
}

func TestRouter_CORSHeaders(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/accounts", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRouter_UnknownRoute(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_GetAccount(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/accounts/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
