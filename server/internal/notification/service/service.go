package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-core-banking-system/internal/notification/domain"
)

type transactionEvent struct {
	TransactionID  string `json:"transaction_id"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	FromAccountID  string `json:"from_account_id"`
	ToAccountID    string `json:"to_account_id"`
	Amount         int64  `json:"amount"`
	IdempotencyKey string `json:"idempotency_key"`
}

type notificationService struct {
	repo domain.Repository
}

func NewNotificationService(repo domain.Repository) domain.Service {
	return &notificationService{repo: repo}
}

func (s *notificationService) HandleTransactionEvent(eventData []byte) error {
	var event transactionEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	notifications := s.createNotifications(event)
	for _, n := range notifications {
		if err := s.repo.Create(n); err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	return nil
}

type target struct {
	account string
	message string
}

func newNotification(txID string, now time.Time, nType domain.NotificationType, t target) *domain.Notification {
	return &domain.Notification{
		ID:            uuid.New().String(),
		Type:          nType,
		TransactionID: txID,
		AccountID:     t.account,
		Message:       t.message,
		Read:          false,
		CreatedAt:     now,
	}
}

func (s *notificationService) createNotifications(event transactionEvent) []*domain.Notification {
	now := time.Now()

	var nType domain.NotificationType
	var targets []target

	switch event.Type {
	case "transaction.completed":
		nType = domain.NotificationTypeTransactionCompleted
		targets = []target{
			{event.FromAccountID, fmt.Sprintf("Transfer of %d completed to %s", event.Amount, event.ToAccountID)},
			{event.ToAccountID, fmt.Sprintf("Received transfer of %d from %s", event.Amount, event.FromAccountID)},
		}
	case "transaction.failed":
		nType = domain.NotificationTypeTransactionFailed
		targets = []target{
			{event.FromAccountID, fmt.Sprintf("Transfer of %d to %s failed", event.Amount, event.ToAccountID)},
		}
	default:
		return nil
	}

	notifications := make([]*domain.Notification, 0, len(targets))
	for _, t := range targets {
		notifications = append(notifications, newNotification(event.TransactionID, now, nType, t))
	}
	return notifications
}

func (s *notificationService) GetByAccountID(accountID string) ([]*domain.Notification, error) {
	return s.repo.GetByAccountID(accountID)
}

func (s *notificationService) MarkAsRead(id string) error {
	return s.repo.MarkAsRead(id)
}
