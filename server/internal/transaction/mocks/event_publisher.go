package mocks

import (
	"go-core-banking-system/internal/transaction/domain"
)

type MockEventPublisher struct {
	PublishTransactionEventFunc func(event *domain.TransactionEvent) error
}

func (m *MockEventPublisher) PublishTransactionEvent(event *domain.TransactionEvent) error {
	if m.PublishTransactionEventFunc != nil {
		return m.PublishTransactionEventFunc(event)
	}
	return nil
}
