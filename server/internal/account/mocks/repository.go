package mocks

import (
	"context"

	"go-core-banking-system/internal/account/domain"
)

type MockAccountRepository struct {
	CreateFunc            func(ctx context.Context, account *domain.Account) error
	GetByIDFunc           func(ctx context.Context, id string) (*domain.Account, error)
	UpdateStatusFunc      func(ctx context.Context, id string, status domain.Status) error
	ReserveFunc           func(ctx context.Context, id string, amount int64) error
	CreditFunc            func(ctx context.Context, id string, amount int64) error
	DebitFunc             func(ctx context.Context, id string, amount int64) error
	CommitReservationFunc func(ctx context.Context, id string, amount int64) error
	CancelReservationFunc func(ctx context.Context, id string, amount int64) error
	BlockAtomicallyFunc   func(ctx context.Context, id string) error
	UnblockAtomicallyFunc func(ctx context.Context, id string) error
	CloseAtomicallyFunc   func(ctx context.Context, id string) error
}

func (m *MockAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, account)
	}
	return nil
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrAccountNotFound
}

func (m *MockAccountRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *MockAccountRepository) Reserve(ctx context.Context, id string, amount int64) error {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountRepository) Credit(ctx context.Context, id string, amount int64) error {
	if m.CreditFunc != nil {
		return m.CreditFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountRepository) Debit(ctx context.Context, id string, amount int64) error {
	if m.DebitFunc != nil {
		return m.DebitFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountRepository) CommitReservation(ctx context.Context, id string, amount int64) error {
	if m.CommitReservationFunc != nil {
		return m.CommitReservationFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountRepository) CancelReservation(ctx context.Context, id string, amount int64) error {
	if m.CancelReservationFunc != nil {
		return m.CancelReservationFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountRepository) BlockAtomically(ctx context.Context, id string) error {
	if m.BlockAtomicallyFunc != nil {
		return m.BlockAtomicallyFunc(ctx, id)
	}
	return nil
}

func (m *MockAccountRepository) UnblockAtomically(ctx context.Context, id string) error {
	if m.UnblockAtomicallyFunc != nil {
		return m.UnblockAtomicallyFunc(ctx, id)
	}
	return nil
}

func (m *MockAccountRepository) CloseAtomically(ctx context.Context, id string) error {
	if m.CloseAtomicallyFunc != nil {
		return m.CloseAtomicallyFunc(ctx, id)
	}
	return nil
}
