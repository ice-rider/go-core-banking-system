package domain

import "time"

type Transaction struct {
	ID             string
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	Status         TransactionStatus
	IdempotencyKey string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TransactionStatus string

const (
	TxStatusPending     TransactionStatus = "PENDING"
	TxStatusCompleted   TransactionStatus = "COMPLETED"
	TxStatusFailed      TransactionStatus = "FAILED"
	TxStatusCompensated TransactionStatus = "COMPENSATED"
)
