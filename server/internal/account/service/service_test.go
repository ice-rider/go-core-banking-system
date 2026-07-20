package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/account/domain"
	"go-core-banking-system/internal/account/mocks"
)

// ==================== Create ====================

func TestCreate_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CreateFunc: func(ctx context.Context, account *domain.Account) error {
			return nil
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: "Alice"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Alice", result.OwnerName)
	assert.Equal(t, int64(0), result.Balance)
	assert.Equal(t, domain.StatusActive, result.Status)
	assert.NotEmpty(t, result.ID)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
}

func TestCreate_EmptyOwnerName(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: ""})

	require.ErrorIs(t, err, domain.ErrOwnerNameRequired)
	assert.Nil(t, result)
}

func TestCreate_RepoError(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CreateFunc: func(ctx context.Context, account *domain.Account) error {
			return errors.New("db connection refused")
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: "Bob"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "db connection refused")
}

// ==================== GetByID ====================

func TestGetByID_Success(t *testing.T) {
	expected := &domain.Account{ID: "acc-1", OwnerName: "Charlie", Balance: 500, Status: domain.StatusActive}
	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			assert.Equal(t, "acc-1", id)
			return expected, nil
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(context.Background(), "acc-1")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(context.Background(), "nonexistent")

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.Nil(t, result)
}

func TestGetByID_RepoGenericError(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, errors.New("timeout")
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(context.Background(), "acc-1")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

// ==================== Block ====================

func TestBlock_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		BlockAtomicallyFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Block(context.Background(), "acc-1")

	require.NoError(t, err)
}

func TestBlock_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		BlockAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Block(context.Background(), "nonexistent")

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestBlock_AlreadyBlocked(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		BlockAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountBlocked
		},
	}
	svc := NewAccountService(repo)

	err := svc.Block(context.Background(), "acc-1")

	require.ErrorIs(t, err, domain.ErrAccountBlocked)
}

// ==================== Unblock ====================

func TestUnblock_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		UnblockAtomicallyFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Unblock(context.Background(), "acc-1")

	require.NoError(t, err)
}

func TestUnblock_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		UnblockAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Unblock(context.Background(), "nonexistent")

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestUnblock_NotBlocked(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		UnblockAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotActive
		},
	}
	svc := NewAccountService(repo)

	err := svc.Unblock(context.Background(), "acc-1")

	require.ErrorIs(t, err, domain.ErrAccountNotActive)
}

// ==================== Close ====================

func TestClose_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			assert.Equal(t, "acc-1", id)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Close(context.Background(), "acc-1")

	require.NoError(t, err)
}

func TestClose_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Close(context.Background(), "nonexistent")

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestClose_BalanceNotZero(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrBalanceNotZero
		},
	}
	svc := NewAccountService(repo)

	err := svc.Close(context.Background(), "acc-1")

	require.ErrorIs(t, err, domain.ErrBalanceNotZero)
}

func TestClose_AccountClosed(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrAccountClosed
		},
	}
	svc := NewAccountService(repo)

	err := svc.Close(context.Background(), "acc-1")

	require.ErrorIs(t, err, domain.ErrAccountClosed)
}

// ==================== Reserve ====================

func TestReserve_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(100), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(context.Background(), "acc-1", 100)

	require.NoError(t, err)
}

func TestReserve_ZeroAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Reserve(context.Background(), "acc-1", 0)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestReserve_NegativeAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Reserve(context.Background(), "acc-1", -50)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestReserve_InsufficientFunds(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(context.Background(), "acc-1", 1000)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestReserve_AccountNotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(context.Background(), "nonexistent", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

// ==================== Credit ====================

func TestCredit_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(200), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Credit(context.Background(), "acc-1", 200)

	require.NoError(t, err)
}

func TestCredit_ZeroAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Credit(context.Background(), "acc-1", 0)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCredit_NegativeAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Credit(context.Background(), "acc-1", -100)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCredit_AccountNotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Credit(context.Background(), "nonexistent", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCredit_AccountBlocked(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountBlocked
		},
	}
	svc := NewAccountService(repo)

	err := svc.Credit(context.Background(), "acc-1", 100)

	require.ErrorIs(t, err, domain.ErrAccountBlocked)
}

// ==================== Debit ====================

func TestDebit_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(150), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Debit(context.Background(), "acc-1", 150)

	require.NoError(t, err)
}

func TestDebit_ZeroAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Debit(context.Background(), "acc-1", 0)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestDebit_NegativeAmount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Debit(context.Background(), "acc-1", -10)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestDebit_InsufficientFunds(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.Debit(context.Background(), "acc-1", 10000)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestDebit_AccountNotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.Debit(context.Background(), "nonexistent", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

// ==================== CommitReservation ====================

func TestCommitReservation_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(300), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.CommitReservation(context.Background(), "acc-1", 300)

	require.NoError(t, err)
}

func TestCommitReservation_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.CommitReservation(context.Background(), "nonexistent", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCommitReservation_InsufficientFunds(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.CommitReservation(context.Background(), "acc-1", 500)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

// ==================== CancelReservation ====================

func TestCancelReservation_Success(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, "acc-1", id)
			assert.Equal(t, int64(250), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.CancelReservation(context.Background(), "acc-1", 250)

	require.NoError(t, err)
}

func TestCancelReservation_NotFound(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.CancelReservation(context.Background(), "nonexistent", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCancelReservation_InsufficientFunds(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.CancelReservation(context.Background(), "acc-1", 1000)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}
