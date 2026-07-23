package mocks

import (
	"go-core-banking-system/internal/notification/domain"
)

type MockNotificationRepository struct {
	CreateFunc         func(notification *domain.Notification) error
	GetByIDFunc        func(id string) (*domain.Notification, error)
	GetByAccountIDFunc func(accountID string) ([]*domain.Notification, error)
	MarkAsReadFunc     func(id string) error
}

func (m *MockNotificationRepository) Create(notification *domain.Notification) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(notification)
	}
	return nil
}

func (m *MockNotificationRepository) GetByID(id string) (*domain.Notification, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockNotificationRepository) GetByAccountID(accountID string) ([]*domain.Notification, error) {
	if m.GetByAccountIDFunc != nil {
		return m.GetByAccountIDFunc(accountID)
	}
	return nil, nil
}

func (m *MockNotificationRepository) MarkAsRead(id string) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(id)
	}
	return nil
}
