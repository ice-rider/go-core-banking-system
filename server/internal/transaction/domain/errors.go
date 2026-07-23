package domain

import "go-core-banking-system/pkg/errors"

var (
	ErrTransactionNotFound    = errors.ErrTransactionNotFound
	ErrIdempotencyKeyUsed     = errors.ErrIdempotencyKeyUsed
	ErrIdempotencyKeyRequired = errors.ErrIdempotencyKeyRequired
	ErrSameAccount            = errors.ErrSameAccount
	ErrAccountNotFound        = errors.ErrAccountNotFound
	ErrAccountBlocked         = errors.ErrAccountBlocked
	ErrInsufficientFunds      = errors.ErrInsufficientFunds
	ErrInvalidAmount          = errors.ErrInvalidAmount
)
