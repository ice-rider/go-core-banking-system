package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func mustUnmarshalHealth(t *testing.T, data []byte) healthResponse {
	t.Helper()
	var resp healthResponse
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

func newTestContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	return c, w
}

func TestHealthCheck_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AccountGRPCHandler{}

	c, w := newTestContext("GET", "/healthz", "")
	h.HealthCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := mustUnmarshalHealth(t, w.Body.Bytes())
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "account-service", resp.Service)
}

func TestHealthCheck_MethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AccountGRPCHandler{}

	c, w := newTestContext("POST", "/healthz", "")
	h.HealthCheck(c)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
