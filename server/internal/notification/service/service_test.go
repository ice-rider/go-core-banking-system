package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/notification/domain"
	"go-core-banking-system/internal/notification/mocks"
)

// ──────────────────────────────────────────────────────
// HandleTransactionEvent
// ──────────────────────────────────────────────────────

func TestHandleTransactionEvent_EmptyBody(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	err := svc.HandleTransactionEvent([]byte(`{}`))

	require.NoError(t, err)
}

func TestHandleTransactionEvent_MissingTypeField(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	event := `{"from_account_id":"acc-from","to_account_id":"acc-to","amount":100}`
	err := svc.HandleTransactionEvent([]byte(event))

	require.NoError(t, err)
}

func TestHandleTransactionEvent_MissingDataField(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	event := `{"type":"transaction.completed"}`
	err := svc.HandleTransactionEvent([]byte(event))

	require.NoError(t, err)
}

func TestHandleTransactionEvent_NestedData(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	evt := transactionEvent{
		TransactionID: "tx-nested",
		Type:          "transaction.completed",
		Status:        "COMPLETED",
		FromAccountID: "acc-from",
		ToAccountID:   "acc-to",
		Amount:        1000,
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 2)
	assert.Equal(t, "tx-nested", created[0].TransactionID)
}

func TestHandleTransactionEvent_LargePayload(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	largeMsg := strings.Repeat("x", 12000)
	evt := map[string]interface{}{
		"transaction_id":  "tx-large",
		"type":            "transaction.completed",
		"status":          "COMPLETED",
		"from_account_id": "acc-from",
		"to_account_id":   "acc-to",
		"amount":          100,
		"idempotency_key": largeMsg,
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 2)
}

func TestHandleTransactionEvent_Concurrent(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var mu sync.Mutex
	var count int
	repo.CreateFunc = func(n *domain.Notification) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}

	evt := transactionEvent{
		TransactionID: "tx-conc",
		Type:          "transaction.completed",
		Status:        "COMPLETED",
		FromAccountID: "acc-from",
		ToAccountID:   "acc-to",
		Amount:        100,
	}
	data, _ := json.Marshal(evt)

	var wg sync.WaitGroup
	errs := make([]error, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = svc.HandleTransactionEvent(data)
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}
	mu.Lock()
	assert.Equal(t, 100, count)
	mu.Unlock()
}

func TestHandleTransactionEvent_Completed_CreatesNotifications(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	evt := transactionEvent{
		TransactionID:  "tx-verify",
		Type:           "transaction.completed",
		Status:         "COMPLETED",
		FromAccountID:  "sender-42",
		ToAccountID:    "receiver-99",
		Amount:         7500,
		IdempotencyKey: "idem-1",
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 2)

	sender := created[0]
	assert.Equal(t, domain.NotificationTypeTransactionCompleted, sender.Type)
	assert.Equal(t, "tx-verify", sender.TransactionID)
	assert.Equal(t, "sender-42", sender.AccountID)
	assert.Contains(t, sender.Message, "7500")
	assert.Contains(t, sender.Message, "receiver-99")
	assert.False(t, sender.Read)
	assert.WithinDuration(t, time.Now(), sender.CreatedAt, time.Second)

	receiver := created[1]
	assert.Equal(t, domain.NotificationTypeTransactionCompleted, receiver.Type)
	assert.Equal(t, "tx-verify", receiver.TransactionID)
	assert.Equal(t, "receiver-99", receiver.AccountID)
	assert.Contains(t, receiver.Message, "7500")
	assert.Contains(t, receiver.Message, "sender-42")
	assert.False(t, receiver.Read)
	assert.WithinDuration(t, time.Now(), receiver.CreatedAt, time.Second)
}

func TestHandleTransactionEvent_Failed_CreatesNotifications(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	evt := transactionEvent{
		TransactionID:  "tx-fail",
		Type:           "transaction.failed",
		Status:         "FAILED",
		FromAccountID:  "sender-10",
		ToAccountID:    "receiver-20",
		Amount:         250,
		IdempotencyKey: "idem-fail",
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 1)

	n := created[0]
	assert.Equal(t, domain.NotificationTypeTransactionFailed, n.Type)
	assert.Equal(t, "tx-fail", n.TransactionID)
	assert.Equal(t, "sender-10", n.AccountID)
	assert.Contains(t, n.Message, "250")
	assert.Contains(t, n.Message, "receiver-20")
	assert.Contains(t, n.Message, "failed")
	assert.False(t, n.Read)
}

func TestHandleTransactionEvent_Completed_WithNegativeAmount(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	evt := transactionEvent{
		TransactionID: "tx-neg",
		Type:          "transaction.completed",
		Status:        "COMPLETED",
		FromAccountID: "acc-a",
		ToAccountID:   "acc-b",
		Amount:        -500,
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 2)
	assert.Contains(t, created[0].Message, "-500")
	assert.Contains(t, created[1].Message, "-500")
}

func TestHandleTransactionEvent_Completed_WithZeroAmount(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var created []*domain.Notification
	repo.CreateFunc = func(n *domain.Notification) error {
		created = append(created, n)
		return nil
	}

	evt := transactionEvent{
		TransactionID: "tx-zero",
		Type:          "transaction.completed",
		Status:        "COMPLETED",
		FromAccountID: "acc-x",
		ToAccountID:   "acc-y",
		Amount:        0,
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.NoError(t, err)
	require.Len(t, created, 2)
	assert.Contains(t, created[0].Message, "0")
	assert.Contains(t, created[1].Message, "0")
}

// ──────────────────────────────────────────────────────
// GetByAccountID
// ──────────────────────────────────────────────────────

func TestGetByAccountID_Success(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	expected := []*domain.Notification{
		{ID: "n-1", AccountID: "acc-1", Message: "msg-1"},
	}
	repo.GetByAccountIDFunc = func(accountID string) ([]*domain.Notification, error) {
		return expected, nil
	}

	result, err := svc.GetByAccountID("acc-1")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetByAccountID_Empty(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.GetByAccountIDFunc = func(accountID string) ([]*domain.Notification, error) {
		return []*domain.Notification{}, nil
	}

	result, err := svc.GetByAccountID("acc-empty")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetByAccountID_RepoError(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.GetByAccountIDFunc = func(accountID string) ([]*domain.Notification, error) {
		return nil, fmt.Errorf("database connection lost")
	}

	result, err := svc.GetByAccountID("acc-err")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection lost")
	assert.Nil(t, result)
}

func TestGetByAccountID_MultipleNotifications(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	expected := make([]*domain.Notification, 50)
	for i := 0; i < 50; i++ {
		expected[i] = &domain.Notification{
			ID:        fmt.Sprintf("n-%d", i),
			AccountID: "acc-many",
			Message:   fmt.Sprintf("message-%d", i),
		}
	}
	repo.GetByAccountIDFunc = func(accountID string) ([]*domain.Notification, error) {
		return expected, nil
	}

	result, err := svc.GetByAccountID("acc-many")

	require.NoError(t, err)
	require.Len(t, result, 50)
	for i, n := range result {
		assert.Equal(t, fmt.Sprintf("n-%d", i), n.ID)
	}
}

// ──────────────────────────────────────────────────────
// MarkAsRead
// ──────────────────────────────────────────────────────

func TestMarkAsRead_Success(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	var markedID string
	repo.MarkAsReadFunc = func(id string) error {
		markedID = id
		return nil
	}

	err := svc.MarkAsRead("n-1")

	require.NoError(t, err)
	assert.Equal(t, "n-1", markedID)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.MarkAsReadFunc = func(id string) error {
		return fmt.Errorf("notification not found")
	}

	err := svc.MarkAsRead("n-nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "notification not found")
}

func TestMarkAsRead_RepoError(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.MarkAsReadFunc = func(id string) error {
		return fmt.Errorf("database timeout")
	}

	err := svc.MarkAsRead("n-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database timeout")
}

func TestMarkAsRead_AlreadyRead(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	repo.MarkAsReadFunc = func(id string) error {
		return nil
	}

	err := svc.MarkAsRead("n-already-read")

	require.NoError(t, err)
}

// ──────────────────────────────────────────────────────
// newNotification
// ──────────────────────────────────────────────────────

func TestNewNotification_FieldsSet(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	tgt := target{account: "acc-test", message: "test message"}

	n := newNotification("tx-1", now, domain.NotificationTypeTransactionCompleted, tgt)

	assert.NotEmpty(t, n.ID)
	assert.Equal(t, domain.NotificationTypeTransactionCompleted, n.Type)
	assert.Equal(t, "tx-1", n.TransactionID)
	assert.Equal(t, "acc-test", n.AccountID)
	assert.Equal(t, "test message", n.Message)
	assert.False(t, n.Read)
	assert.Equal(t, now, n.CreatedAt)
}

func TestNewNotification_DefaultValues(t *testing.T) {
	n := newNotification("tx-2", time.Now(), domain.NotificationTypeTransactionFailed, target{})

	assert.False(t, n.Read)
	assert.NotEmpty(t, n.ID)
}

// ──────────────────────────────────────────────────────
// createNotifications
// ──────────────────────────────────────────────────────

func TestCreateNotifications_MultipleEvents(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo).(*notificationService)

	evt := transactionEvent{
		TransactionID: "tx-multi",
		Type:          "transaction.completed",
		FromAccountID: "acc-1",
		ToAccountID:   "acc-2",
		Amount:        500,
	}
	notifications := svc.createNotifications(evt)

	require.Len(t, notifications, 2)
	assert.Equal(t, "acc-1", notifications[0].AccountID)
	assert.Equal(t, "acc-2", notifications[1].AccountID)
	assert.Equal(t, domain.NotificationTypeTransactionCompleted, notifications[0].Type)
}

func TestCreateNotifications_RepoError(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo)

	callCount := 0
	repo.CreateFunc = func(n *domain.Notification) error {
		callCount++
		if callCount == 1 {
			return fmt.Errorf("repo write error")
		}
		return nil
	}

	evt := transactionEvent{
		TransactionID: "tx-err",
		Type:          "transaction.completed",
		FromAccountID: "acc-from",
		ToAccountID:   "acc-to",
		Amount:        100,
	}
	data, _ := json.Marshal(evt)

	err := svc.HandleTransactionEvent(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create notification")
}

func TestCreateNotifications_FailedType(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo).(*notificationService)

	evt := transactionEvent{
		TransactionID: "tx-fail-multi",
		Type:          "transaction.failed",
		FromAccountID: "acc-a",
		ToAccountID:   "acc-b",
		Amount:        200,
	}
	notifications := svc.createNotifications(evt)

	require.Len(t, notifications, 1)
	assert.Equal(t, domain.NotificationTypeTransactionFailed, notifications[0].Type)
	assert.Equal(t, "acc-a", notifications[0].AccountID)
}

func TestCreateNotifications_UnknownType(t *testing.T) {
	repo := &mocks.MockNotificationRepository{}
	svc := NewNotificationService(repo).(*notificationService)

	evt := transactionEvent{
		Type: "unknown.event",
	}
	notifications := svc.createNotifications(evt)

	assert.Nil(t, notifications)
}
