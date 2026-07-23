package errors

import "errors"

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrAccountBlocked    = errors.New("account is blocked")
	ErrAccountClosed     = errors.New("account is closed")
	ErrAccountNotActive  = errors.New("account is not active")
	ErrBalanceNotZero    = errors.New("balance must be zero to close account")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrOwnerNameRequired = errors.New("owner name is required")

	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrIdempotencyKeyUsed     = errors.New("idempotency key already used")
	ErrIdempotencyKeyRequired = errors.New("idempotency key is required")
	ErrSameAccount            = errors.New("cannot transfer to the same account")
)

var sentinelErrors = []error{
	ErrAccountNotFound,
	ErrAccountBlocked,
	ErrAccountClosed,
	ErrAccountNotActive,
	ErrBalanceNotZero,
	ErrInsufficientFunds,
	ErrInvalidAmount,
	ErrOwnerNameRequired,
	ErrTransactionNotFound,
	ErrIdempotencyKeyUsed,
	ErrIdempotencyKeyRequired,
	ErrSameAccount,
}

func IsDomainError(err error) bool {
	if err == nil {
		return false
	}
	for _, sentinel := range sentinelErrors {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}

func ToHTTPStatus(err error) (int, string) {
	if err == nil {
		return 200, ""
	}

	switch {
	case errors.Is(err, ErrAccountNotFound), errors.Is(err, ErrTransactionNotFound):
		return 404, "NOT_FOUND"
	case errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrOwnerNameRequired),
		errors.Is(err, ErrIdempotencyKeyRequired), errors.Is(err, ErrSameAccount):
		return 400, "VALIDATION_ERROR"
	case errors.Is(err, ErrInsufficientFunds), errors.Is(err, ErrAccountBlocked),
		errors.Is(err, ErrAccountClosed), errors.Is(err, ErrAccountNotActive),
		errors.Is(err, ErrBalanceNotZero), errors.Is(err, ErrIdempotencyKeyUsed):
		return 409, "CONFLICT"
	default:
		return 500, "INTERNAL_ERROR"
	}
}
