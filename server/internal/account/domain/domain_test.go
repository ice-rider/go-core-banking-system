package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{"active value", StatusActive, "ACTIVE"},
		{"blocked value", StatusBlocked, "BLOCKED"},
		{"closed value", StatusClosed, "CLOSED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, Status(tt.expected), tt.status)
		})
	}
}

func TestStatusAreDistinct(t *testing.T) {
	statuses := []Status{StatusActive, StatusBlocked, StatusClosed}
	seen := make(map[Status]bool)
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
		{"ErrAccountNotFound", ErrAccountNotFound, "account not found"},
		{"ErrInsufficientFunds", ErrInsufficientFunds, "insufficient funds"},
		{"ErrAccountBlocked", ErrAccountBlocked, "account is blocked"},
		{"ErrAccountClosed", ErrAccountClosed, "account is closed"},
		{"ErrAccountNotActive", ErrAccountNotActive, "account is not active"},
		{"ErrBalanceNotZero", ErrBalanceNotZero, "balance must be zero to close account"},
		{"ErrInvalidAmount", ErrInvalidAmount, "amount must be positive"},
		{"ErrOwnerNameRequired", ErrOwnerNameRequired, "owner name is required"},
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
		ErrAccountNotFound,
		ErrInsufficientFunds,
		ErrAccountBlocked,
		ErrAccountClosed,
		ErrAccountNotActive,
		ErrBalanceNotZero,
		ErrInvalidAmount,
		ErrOwnerNameRequired,
	}
	for _, err := range errs {
		assert.NotNil(t, err)
	}
}
