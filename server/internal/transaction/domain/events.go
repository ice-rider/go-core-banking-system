package domain

import "time"

type EventType string

const (
	EventTransactionCompleted EventType = "transaction.completed"
	EventTransactionFailed    EventType = "transaction.failed"
)

type TransactionEvent struct {
	ID             string            `json:"id"`
	TransactionID  string            `json:"transaction_id"`
	Type           EventType         `json:"type"`
	Status         TransactionStatus `json:"status"`
	FromAccountID  string            `json:"from_account_id"`
	ToAccountID    string            `json:"to_account_id"`
	Amount         int64             `json:"amount"`
	IdempotencyKey string            `json:"idempotency_key"`
	Timestamp      time.Time         `json:"timestamp"`
}

type EventPublisher interface {
	PublishTransactionEvent(event *TransactionEvent) error
}
