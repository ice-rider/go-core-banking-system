package domain

import "go-core-banking-system/pkg/errors"

var (
	ErrAccountNotFound   = errors.ErrAccountNotFound
	ErrInsufficientFunds = errors.ErrInsufficientFunds
	ErrAccountBlocked    = errors.ErrAccountBlocked
	ErrAccountClosed     = errors.ErrAccountClosed
	ErrAccountNotActive  = errors.ErrAccountNotActive
	ErrBalanceNotZero    = errors.ErrBalanceNotZero
	ErrInvalidAmount     = errors.ErrInvalidAmount
	ErrOwnerNameRequired = errors.ErrOwnerNameRequired
)
