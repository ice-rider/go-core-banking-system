package domain

import "errors"

var (
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrIdempotencyKeyUsed       = errors.New("idempotency key already used")
	ErrIdempotencyKeyRequired   = errors.New("idempotency key is required")
	ErrSameAccount              = errors.New("cannot transfer to the same account")
	ErrAccountNotFound          = errors.New("account not found")
	ErrAccountBlocked           = errors.New("account is blocked")
	ErrInsufficientFunds        = errors.New("insufficient funds")
	ErrInvalidAmount            = errors.New("amount must be positive")
)
