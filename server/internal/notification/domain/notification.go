package domain

import "time"

type NotificationType string

const (
	NotificationTypeTransactionCompleted NotificationType = "transaction_completed"
	NotificationTypeTransactionFailed    NotificationType = "transaction_failed"
)

type Notification struct {
	ID            string           `json:"id"`
	Type          NotificationType `json:"type"`
	TransactionID string           `json:"transaction_id"`
	AccountID     string           `json:"account_id"`
	Message       string           `json:"message"`
	Read          bool             `json:"read"`
	CreatedAt     time.Time        `json:"created_at"`
}

type Repository interface {
	Create(notification *Notification) error
	GetByID(id string) (*Notification, error)
	GetByAccountID(accountID string) ([]*Notification, error)
	MarkAsRead(id string) error
}

type Service interface {
	HandleTransactionEvent(eventData []byte) error
	GetByAccountID(accountID string) ([]*Notification, error)
	MarkAsRead(id string) error
}
