package domain

import "errors"

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrAccountBlocked       = errors.New("account is blocked")
	ErrAccountClosed        = errors.New("account is closed")
	ErrAccountNotActive     = errors.New("account is not active")
	ErrBalanceNotZero       = errors.New("balance must be zero to close account")
	ErrInvalidAmount        = errors.New("amount must be positive")
	ErrOwnerNameRequired    = errors.New("owner name is required")
)
