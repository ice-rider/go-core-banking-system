package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/notification/domain"
	"go-core-banking-system/internal/notification/mocks"
)

func TestHandleTransactionEvent_Completed(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var createdNotifications []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		createdNotifications = append(createdNotifications, n)
		return nil
	}

	event := transactionEvent{
		TransactionID:  "tx-1",
		Type:           "transaction.completed",
		Status:         "COMPLETED",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         500,
		IdempotencyKey: "key-1",
	}
	eventData, _ := json.Marshal(event)

	err := svc.HandleTransactionEvent(eventData)

	require.NoError(t, err)
	require.Len(t, createdNotifications, 2)
	assert.Equal(t, domain.NotificationTypeTransactionCompleted, createdNotifications[0].Type)
	assert.Equal(t, "acc-from", createdNotifications[0].AccountID)
	assert.Equal(t, "acc-to", createdNotifications[1].AccountID)
}

func TestHandleTransactionEvent_Failed(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var createdNotifications []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		createdNotifications = append(createdNotifications, n)
		return nil
	}

	event := transactionEvent{
		TransactionID:  "tx-2",
		Type:           "transaction.failed",
		Status:         "FAILED",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         300,
		IdempotencyKey: "key-2",
	}
	eventData, _ := json.Marshal(event)

	err := svc.HandleTransactionEvent(eventData)

	require.NoError(t, err)
	require.Len(t, createdNotifications, 1)
	assert.Equal(t, domain.NotificationTypeTransactionFailed, createdNotifications[0].Type)
	assert.Equal(t, "acc-from", createdNotifications[0].AccountID)
}

func TestHandleTransactionEvent_InvalidJSON(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	err := svc.HandleTransactionEvent([]byte("invalid json"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal event")
}

func TestHandleTransactionEvent_RepoError(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.CreateFunc = func(n *domain.Notification) error {
		return assert.AnError
	}

	event := transactionEvent{
		TransactionID:  "tx-3",
		Type:           "transaction.completed",
		Status:         "COMPLETED",
		FromAccountID:  "acc-from",
		ToAccountID:    "acc-to",
		Amount:         100,
		IdempotencyKey: "key-3",
	}
	eventData, _ := json.Marshal(event)

	err := svc.HandleTransactionEvent(eventData)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create notification")
}

func TestGetByAccountID_Success(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	expected := []*domain.Notification{
		{ID: "n-1", AccountID: "acc-1", Message: "test"},
	}
	repo.GetByAccountIDFunc = func(accountID string) ([]*domain.Notification, error) {
		return expected, nil
	}

	result, err := svc.GetByAccountID("acc-1")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestHandleTransactionEvent_UnknownType(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	event := `{"type":"unknown.event","data":{}}`
	err := svc.HandleTransactionEvent([]byte(event))

	require.NoError(t, err)
}

func TestMarkAsRead_Success(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.MarkAsReadFunc = func(id string) error {
		return nil
	}

	err := svc.MarkAsRead("n-1")

	require.NoError(t, err)
}
