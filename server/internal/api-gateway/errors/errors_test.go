package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	// ==================== Nil error ====================
	t.Run("nil_error", func(t *testing.T) {
		status, resp := MapError(nil)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, ErrorResponse{}, resp)
	})

	// ==================== 404 Not Found ====================
	t.Run("account_not_found", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("account not found"))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
		assert.Contains(t, resp.Error, "account not found")
	})

	t.Run("transaction_not_found", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("transaction not found"))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
		assert.Contains(t, resp.Error, "transaction not found")
	})

	t.Run("not_found_wrapped_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("get account: %w", fmt.Errorf("account not found")))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
	})

	// ==================== 400 Validation Error ====================
	t.Run("amount_must_be_positive", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("amount must be positive"))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("owner_name_required", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("owner name is required"))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("idempotency_key_required", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("idempotency key is required"))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("same_account_transfer", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("cannot transfer to the same account"))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	// ==================== 409 Conflict ====================
	t.Run("insufficient_funds", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("insufficient funds"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_blocked", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("account is blocked"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_closed", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("account is closed"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_not_active", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("account is not active"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("balance_not_zero", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("balance must be zero to close account"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("idempotency_key_used", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("idempotency key already used"))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("conflict_wrapped_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("transfer: %w", fmt.Errorf("insufficient funds")))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	// ==================== 500 Internal Error ====================
	t.Run("unknown_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("something went wrong"))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
		assert.Equal(t, "something went wrong", resp.Error)
	})

	t.Run("database_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("connection refused"))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "INTERNAL_ERROR", resp.Code)
	})

	// ==================== ErrorResponse JSON tags ====================
	t.Run("error_response_fields", func(t *testing.T) {
		_, resp := MapError(fmt.Errorf("account not found"))
		require.NotEmpty(t, resp.Error)
		require.NotEmpty(t, resp.Code)
	})
}
