package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToHTTPStatus_Nil(t *testing.T) {
	code, msg := ToHTTPStatus(nil)
	require.Equal(t, 200, code)
	assert.Equal(t, "", msg)
}

func TestToHTTPStatus_NotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrAccountNotFound", ErrAccountNotFound},
		{"ErrTransactionNotFound", ErrTransactionNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := ToHTTPStatus(tt.err)
			assert.Equal(t, 404, code)
			assert.Equal(t, "NOT_FOUND", msg)
		})
	}
}

func TestToHTTPStatus_ValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidAmount", ErrInvalidAmount},
		{"ErrOwnerNameRequired", ErrOwnerNameRequired},
		{"ErrIdempotencyKeyRequired", ErrIdempotencyKeyRequired},
		{"ErrSameAccount", ErrSameAccount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := ToHTTPStatus(tt.err)
			assert.Equal(t, 400, code)
			assert.Equal(t, "VALIDATION_ERROR", msg)
		})
	}
}

func TestToHTTPStatus_Conflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInsufficientFunds", ErrInsufficientFunds},
		{"ErrAccountBlocked", ErrAccountBlocked},
		{"ErrAccountClosed", ErrAccountClosed},
		{"ErrAccountNotActive", ErrAccountNotActive},
		{"ErrBalanceNotZero", ErrBalanceNotZero},
		{"ErrIdempotencyKeyUsed", ErrIdempotencyKeyUsed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := ToHTTPStatus(tt.err)
			assert.Equal(t, 409, code)
			assert.Equal(t, "CONFLICT", msg)
		})
	}
}

func TestToHTTPStatus_InternalError(t *testing.T) {
	code, msg := ToHTTPStatus(fmt.Errorf("unknown error"))
	assert.Equal(t, 500, code)
	assert.Equal(t, "INTERNAL_ERROR", msg)
}

func TestToHTTPStatus_WrappedDomainError(t *testing.T) {
	wrapped := fmt.Errorf("deposited context: %w", ErrAccountNotFound)
	code, msg := ToHTTPStatus(wrapped)
	assert.Equal(t, 404, code)
	assert.Equal(t, "NOT_FOUND", msg)
}
