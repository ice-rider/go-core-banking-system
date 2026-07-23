package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/account/domain"
	"go-core-banking-system/pkg/proto/account"
)

type mockAccountService struct {
	CreateFunc            func(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error)
	GetByIDFunc           func(ctx context.Context, id string) (*domain.Account, error)
	BlockFunc             func(ctx context.Context, id string) error
	UnblockFunc           func(ctx context.Context, id string) error
	CloseFunc             func(ctx context.Context, id string) error
	ReserveFunc           func(ctx context.Context, id string, amount int64) error
	CreditFunc            func(ctx context.Context, id string, amount int64) error
	DebitFunc             func(ctx context.Context, id string, amount int64) error
	CommitReservationFunc func(ctx context.Context, id string, amount int64) error
	CancelReservationFunc func(ctx context.Context, id string, amount int64) error
}

func (m *mockAccountService) Create(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockAccountService) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAccountService) Block(ctx context.Context, id string) error {
	if m.BlockFunc != nil {
		return m.BlockFunc(ctx, id)
	}
	return nil
}

func (m *mockAccountService) Unblock(ctx context.Context, id string) error {
	if m.UnblockFunc != nil {
		return m.UnblockFunc(ctx, id)
	}
	return nil
}

func (m *mockAccountService) Close(ctx context.Context, id string) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id)
	}
	return nil
}

func (m *mockAccountService) Reserve(ctx context.Context, id string, amount int64) error {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, id, amount)
	}
	return nil
}

func (m *mockAccountService) Credit(ctx context.Context, id string, amount int64) error {
	if m.CreditFunc != nil {
		return m.CreditFunc(ctx, id, amount)
	}
	return nil
}

func (m *mockAccountService) Debit(ctx context.Context, id string, amount int64) error {
	if m.DebitFunc != nil {
		return m.DebitFunc(ctx, id, amount)
	}
	return nil
}

func (m *mockAccountService) CommitReservation(ctx context.Context, id string, amount int64) error {
	if m.CommitReservationFunc != nil {
		return m.CommitReservationFunc(ctx, id, amount)
	}
	return nil
}

func (m *mockAccountService) CancelReservation(ctx context.Context, id string, amount int64) error {
	if m.CancelReservationFunc != nil {
		return m.CancelReservationFunc(ctx, id, amount)
	}
	return nil
}

func testTime() time.Time {
	return time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
}

func testAccount() *domain.Account {
	return &domain.Account{
		ID:        "acc-1",
		OwnerName: "Alice",
		Balance:   1000,
		Status:    domain.StatusActive,
		CreatedAt: testTime(),
		UpdatedAt: testTime(),
	}
}

func TestCreate_Success(t *testing.T) {
	svc := &mockAccountService{
		CreateFunc: func(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error) {
			assert.Equal(t, "Alice", input.OwnerName)
			return testAccount(), nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Create(context.Background(), &account.CreateAccountRequest{OwnerName: "Alice"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "acc-1", resp.Id)
	assert.Equal(t, "Alice", resp.OwnerName)
	assert.Equal(t, int64(1000), resp.Balance)
	assert.Equal(t, "ACTIVE", resp.Status)
	assert.NotEmpty(t, resp.CreatedAt)
	assert.NotEmpty(t, resp.UpdatedAt)
}

func TestCreate_ServiceError(t *testing.T) {
	svc := &mockAccountService{
		CreateFunc: func(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error) {
			return nil, domain.ErrOwnerNameRequired
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Create(context.Background(), &account.CreateAccountRequest{OwnerName: ""})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrOwnerNameRequired)
}

func TestGetByID_Success(t *testing.T) {
	svc := &mockAccountService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			assert.Equal(t, "acc-1", id)
			return testAccount(), nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &account.GetAccountRequest{Id: "acc-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "acc-1", resp.Id)
	assert.Equal(t, "Alice", resp.OwnerName)
	assert.Equal(t, int64(1000), resp.Balance)
	assert.Equal(t, "ACTIVE", resp.Status)
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockAccountService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &account.GetAccountRequest{Id: "nonexistent"})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestGetByID_RepoError(t *testing.T) {
	svc := &mockAccountService{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, errors.New("db timeout")
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.GetByID(context.Background(), &account.GetAccountRequest{Id: "acc-1"})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "db timeout")
}

func TestBlock_Success(t *testing.T) {
	svc := &mockAccountService{
		BlockFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Block(context.Background(), &account.BlockAccountRequest{Id: "acc-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestBlock_NotFound(t *testing.T) {
	svc := &mockAccountService{
		BlockFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Block(context.Background(), &account.BlockAccountRequest{Id: "nonexistent"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestBlock_AlreadyBlocked(t *testing.T) {
	svc := &mockAccountService{
		BlockFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountBlocked
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Block(context.Background(), &account.BlockAccountRequest{Id: "acc-1"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountBlocked)
}

func TestUnblock_Success(t *testing.T) {
	svc := &mockAccountService{
		UnblockFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Unblock(context.Background(), &account.UnblockAccountRequest{Id: "acc-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestUnblock_NotFound(t *testing.T) {
	svc := &mockAccountService{
		UnblockFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Unblock(context.Background(), &account.UnblockAccountRequest{Id: "nonexistent"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestUnblock_NotBlocked(t *testing.T) {
	svc := &mockAccountService{
		UnblockFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotActive
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Unblock(context.Background(), &account.UnblockAccountRequest{Id: "acc-1"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotActive)
}

func TestClose_Success(t *testing.T) {
	svc := &mockAccountService{
		CloseFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Close(context.Background(), &account.CloseAccountRequest{Id: "acc-1"})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestClose_NotFound(t *testing.T) {
	svc := &mockAccountService{
		CloseFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Close(context.Background(), &account.CloseAccountRequest{Id: "nonexistent"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestClose_BalanceNotZero(t *testing.T) {
	svc := &mockAccountService{
		CloseFunc: func(ctx context.Context, id string) error {
			return domain.ErrBalanceNotZero
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Close(context.Background(), &account.CloseAccountRequest{Id: "acc-1"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBalanceNotZero)
}

func TestClose_AccountClosed(t *testing.T) {
	svc := &mockAccountService{
		CloseFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountClosed
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Close(context.Background(), &account.CloseAccountRequest{Id: "acc-1"})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountClosed)
}

func TestReserve_Success(t *testing.T) {
	svc := &mockAccountService{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(500), amount)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Reserve(context.Background(), &account.ReserveRequest{Id: "acc-1", Amount: 500})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestReserve_InvalidAmount(t *testing.T) {
	svc := &mockAccountService{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInvalidAmount
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Reserve(context.Background(), &account.ReserveRequest{Id: "acc-1", Amount: 0})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestReserve_InsufficientFunds(t *testing.T) {
	svc := &mockAccountService{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Reserve(context.Background(), &account.ReserveRequest{Id: "acc-1", Amount: 10000})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestReserve_NotFound(t *testing.T) {
	svc := &mockAccountService{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Reserve(context.Background(), &account.ReserveRequest{Id: "nonexistent", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCredit_Success(t *testing.T) {
	svc := &mockAccountService{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(200), amount)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Credit(context.Background(), &account.CreditRequest{Id: "acc-1", Amount: 200})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestCredit_InvalidAmount(t *testing.T) {
	svc := &mockAccountService{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInvalidAmount
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Credit(context.Background(), &account.CreditRequest{Id: "acc-1", Amount: -100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCredit_AccountNotFound(t *testing.T) {
	svc := &mockAccountService{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Credit(context.Background(), &account.CreditRequest{Id: "nonexistent", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCredit_AccountBlocked(t *testing.T) {
	svc := &mockAccountService{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountBlocked
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Credit(context.Background(), &account.CreditRequest{Id: "acc-1", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountBlocked)
}

func TestDebit_Success(t *testing.T) {
	svc := &mockAccountService{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(150), amount)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.Debit(context.Background(), &account.DebitRequest{Id: "acc-1", Amount: 150})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestDebit_InvalidAmount(t *testing.T) {
	svc := &mockAccountService{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInvalidAmount
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Debit(context.Background(), &account.DebitRequest{Id: "acc-1", Amount: -10})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestDebit_InsufficientFunds(t *testing.T) {
	svc := &mockAccountService{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Debit(context.Background(), &account.DebitRequest{Id: "acc-1", Amount: 10000})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestDebit_NotFound(t *testing.T) {
	svc := &mockAccountService{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.Debit(context.Background(), &account.DebitRequest{Id: "nonexistent", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCommitReservation_Success(t *testing.T) {
	svc := &mockAccountService{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(300), amount)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.CommitReservation(context.Background(), &account.CommitReservationRequest{Id: "acc-1", Amount: 300})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestCommitReservation_NotFound(t *testing.T) {
	svc := &mockAccountService{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.CommitReservation(context.Background(), &account.CommitReservationRequest{Id: "nonexistent", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCommitReservation_InsufficientFunds(t *testing.T) {
	svc := &mockAccountService{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.CommitReservation(context.Background(), &account.CommitReservationRequest{Id: "acc-1", Amount: 500})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestCancelReservation_Success(t *testing.T) {
	svc := &mockAccountService{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(250), amount)
			return nil
		},
	}
	h := NewAccountGRPCHandler(svc)

	resp, err := h.CancelReservation(context.Background(), &account.CancelReservationRequest{Id: "acc-1", Amount: 250})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.IsType(t, &account.Empty{}, resp)
}

func TestCancelReservation_NotFound(t *testing.T) {
	svc := &mockAccountService{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.CancelReservation(context.Background(), &account.CancelReservationRequest{Id: "nonexistent", Amount: 100})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCancelReservation_InsufficientFunds(t *testing.T) {
	svc := &mockAccountService{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	h := NewAccountGRPCHandler(svc)

	_, err := h.CancelReservation(context.Background(), &account.CancelReservationRequest{Id: "acc-1", Amount: 1000})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestNewAccountGRPCHandler(t *testing.T) {
	svc := &mockAccountService{}
	h := NewAccountGRPCHandler(svc)

	require.NotNil(t, h)
	assert.Equal(t, svc, h.svc)
}
