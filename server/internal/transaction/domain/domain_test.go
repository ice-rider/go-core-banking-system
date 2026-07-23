package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   TransactionStatus
		expected string
	}{
		{"pending value", TxStatusPending, "PENDING"},
		{"completed value", TxStatusCompleted, "COMPLETED"},
		{"failed value", TxStatusFailed, "FAILED"},
		{"compensated value", TxStatusCompensated, "COMPENSATED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, TransactionStatus(tt.expected), tt.status)
		})
	}
}

func TestTransactionStatusAreDistinct(t *testing.T) {
	statuses := []TransactionStatus{
		TxStatusPending,
		TxStatusCompleted,
		TxStatusFailed,
		TxStatusCompensated,
	}
	seen := make(map[TransactionStatus]bool)
	for _, s := range statuses {
		require.False(t, seen[s], "duplicate status: %s", s)
		seen[s] = true
	}
}

func TestDomainErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrTransactionNotFound", ErrTransactionNotFound, "transaction not found"},
		{"ErrIdempotencyKeyUsed", ErrIdempotencyKeyUsed, "idempotency key already used"},
		{"ErrIdempotencyKeyRequired", ErrIdempotencyKeyRequired, "idempotency key is required"},
		{"ErrSameAccount", ErrSameAccount, "cannot transfer to the same account"},
		{"ErrAccountNotFound", ErrAccountNotFound, "account not found"},
		{"ErrAccountBlocked", ErrAccountBlocked, "account is blocked"},
		{"ErrInsufficientFunds", ErrInsufficientFunds, "insufficient funds"},
		{"ErrInvalidAmount", ErrInvalidAmount, "amount must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestDomainErrorsAreNonNil(t *testing.T) {
	errs := []error{
		ErrTransactionNotFound,
		ErrIdempotencyKeyUsed,
		ErrIdempotencyKeyRequired,
		ErrSameAccount,
		ErrAccountNotFound,
		ErrAccountBlocked,
		ErrInsufficientFunds,
		ErrInvalidAmount,
	}
	for _, err := range errs {
		assert.NotNil(t, err)
	}
}
