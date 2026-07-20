package mocks

import (
	"context"
)

type MockAccountClient struct {
	ReserveFunc           func(ctx context.Context, id string, amount int64) error
	CreditFunc            func(ctx context.Context, id string, amount int64) error
	DebitFunc             func(ctx context.Context, id string, amount int64) error
	CommitReservationFunc func(ctx context.Context, id string, amount int64) error
	CancelReservationFunc func(ctx context.Context, id string, amount int64) error
}

func (m *MockAccountClient) Reserve(ctx context.Context, id string, amount int64) error {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountClient) Credit(ctx context.Context, id string, amount int64) error {
	if m.CreditFunc != nil {
		return m.CreditFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountClient) Debit(ctx context.Context, id string, amount int64) error {
	if m.DebitFunc != nil {
		return m.DebitFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountClient) CommitReservation(ctx context.Context, id string, amount int64) error {
	if m.CommitReservationFunc != nil {
		return m.CommitReservationFunc(ctx, id, amount)
	}
	return nil
}

func (m *MockAccountClient) CancelReservation(ctx context.Context, id string, amount int64) error {
	if m.CancelReservationFunc != nil {
		return m.CancelReservationFunc(ctx, id, amount)
	}
	return nil
}
