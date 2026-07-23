package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgerrors "go-core-banking-system/pkg/errors"
)

func TestMapError(t *testing.T) {
	t.Run("nil_error", func(t *testing.T) {
		status, resp := MapError(nil)
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, ErrorResponse{}, resp)
	})

	t.Run("account_not_found", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrAccountNotFound)
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
		assert.Contains(t, resp.Error, "account not found")
	})

	t.Run("transaction_not_found", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrTransactionNotFound)
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
		assert.Contains(t, resp.Error, "transaction not found")
	})

	t.Run("not_found_wrapped_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("get account: %w", pkgerrors.ErrAccountNotFound))
		assert.Equal(t, http.StatusNotFound, status)
		assert.Equal(t, "NOT_FOUND", resp.Code)
	})

	t.Run("amount_must_be_positive", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrInvalidAmount)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("owner_name_required", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrOwnerNameRequired)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("idempotency_key_required", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrIdempotencyKeyRequired)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("same_account_transfer", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrSameAccount)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "VALIDATION_ERROR", resp.Code)
	})

	t.Run("insufficient_funds", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrInsufficientFunds)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_blocked", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrAccountBlocked)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_closed", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrAccountClosed)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("account_not_active", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrAccountNotActive)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("balance_not_zero", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrBalanceNotZero)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("idempotency_key_used", func(t *testing.T) {
		status, resp := MapError(pkgerrors.ErrIdempotencyKeyUsed)
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

	t.Run("conflict_wrapped_error", func(t *testing.T) {
		status, resp := MapError(fmt.Errorf("transfer: %w", pkgerrors.ErrInsufficientFunds))
		assert.Equal(t, http.StatusConflict, status)
		assert.Equal(t, "CONFLICT", resp.Code)
	})

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

	t.Run("error_response_fields", func(t *testing.T) {
		_, resp := MapError(pkgerrors.ErrAccountNotFound)
		require.NotEmpty(t, resp.Error)
		require.NotEmpty(t, resp.Code)
	})
}

func TestIsDomainError_True(t *testing.T) {
	err := pkgerrors.ErrAccountNotFound
	assert.True(t, IsDomainError(err))
}

func TestIsDomainError_False(t *testing.T) {
	err := errors.New("some random error")
	assert.False(t, IsDomainError(err))
}
