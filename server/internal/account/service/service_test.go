package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

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

// ==================== Create — Edge Cases ====================

func TestCreate_ConcurrentCalls(t *testing.T) {
	const goroutines = 50
	var mu sync.Mutex
	createdAccounts := make([]*domain.Account, 0, goroutines)

	repo := &mocks.MockAccountRepository{
		CreateFunc: func(ctx context.Context, account *domain.Account) error {
			mu.Lock()
			createdAccounts = append(createdAccounts, account)
			mu.Unlock()
			return nil
		},
	}
	svc := NewAccountService(repo)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: "Concurrent"})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, int64(0), result.Balance)
			assert.Equal(t, domain.StatusActive, result.Status)
		}()
	}

	wg.Wait()
	require.Len(t, createdAccounts, goroutines)

	ids := make(map[string]bool)
	for _, acc := range createdAccounts {
		assert.False(t, ids[acc.ID], "duplicate ID generated: %s", acc.ID)
		ids[acc.ID] = true
	}
}

func TestCreate_VeryLongOwnerName(t *testing.T) {
	t.Run("255 chars — max valid", func(t *testing.T) {
		name := strings.Repeat("A", 255)
		repo := &mocks.MockAccountRepository{
			CreateFunc: func(ctx context.Context, account *domain.Account) error {
				assert.Equal(t, name, account.OwnerName)
				return nil
			},
		}
		svc := NewAccountService(repo)

		result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: name})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, name, result.OwnerName)
	})

	t.Run("256 chars — over max", func(t *testing.T) {
		name := strings.Repeat("A", 256)
		repo := &mocks.MockAccountRepository{
			CreateFunc: func(ctx context.Context, account *domain.Account) error {
				return nil
			},
		}
		svc := NewAccountService(repo)

		result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: name})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, name, result.OwnerName)
	})
}

func TestCreate_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name      string
		ownerName string
	}{
		{"unicode cyrillic", "Иван Иванов"},
		{"unicode chinese", "张三"},
		{"emoji", "John😀Doe"},
		{"sql injection", "'; DROP TABLE accounts;--"},
		{"sql injection 2", "Robert'); --"},
		{"zero width chars", "John\u200BDoe"},
		{"newlines and tabs", "John\nDoe\tTab"},
		{"quotes", `"John" 'Doe'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedName string
			repo := &mocks.MockAccountRepository{
				CreateFunc: func(ctx context.Context, account *domain.Account) error {
					capturedName = account.OwnerName
					return nil
				},
			}
			svc := NewAccountService(repo)

			result, err := svc.Create(context.Background(), domain.CreateAccountInput{OwnerName: tt.ownerName})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.ownerName, capturedName)
		})
	}
}

// ==================== GetByID — Edge Cases ====================

func TestGetByID_EmptyString(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			assert.Equal(t, "", id)
			return nil, domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(context.Background(), "")

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.Nil(t, result)
}

func TestGetByID_VeryLongID(t *testing.T) {
	longID := strings.Repeat("a", 10000)
	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			assert.Equal(t, longID, id)
			return nil, domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(context.Background(), longID)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.Nil(t, result)
}

func TestGetByID_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mocks.MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id string) (*domain.Account, error) {
			return nil, ctx.Err()
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.GetByID(ctx, "acc-1")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

// ==================== Block/Unblock/Close — Edge Cases ====================

func TestBlock_ConcurrentBlock(t *testing.T) {
	const goroutines = 20
	callCount := 0
	var mu sync.Mutex

	repo := &mocks.MockAccountRepository{
		BlockAtomicallyFunc: func(ctx context.Context, id string) error {
			mu.Lock()
			defer mu.Unlock()
			callCount++
			if callCount > 1 {
				return domain.ErrAccountBlocked
			}
			return nil
		},
	}
	svc := NewAccountService(repo)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = svc.Block(context.Background(), "acc-1")
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}

	require.LessOrEqual(t, successCount, 1, "at most one block should succeed")
}

func TestUnblock_ConcurrentUnblock(t *testing.T) {
	const goroutines = 20
	callCount := 0
	var mu sync.Mutex

	repo := &mocks.MockAccountRepository{
		UnblockAtomicallyFunc: func(ctx context.Context, id string) error {
			mu.Lock()
			defer mu.Unlock()
			callCount++
			if callCount > 1 {
				return domain.ErrAccountNotActive
			}
			return nil
		},
	}
	svc := NewAccountService(repo)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = svc.Unblock(context.Background(), "acc-1")
		}(i)
	}

	wg.Wait()

	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}

	require.LessOrEqual(t, successCount, 1, "at most one unblock should succeed")
}

func TestClose_DuringActiveReservation(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			return domain.ErrBalanceNotZero
		},
	}
	svc := NewAccountService(repo)

	err := svc.Close(context.Background(), "acc-1")

	require.ErrorIs(t, err, domain.ErrBalanceNotZero)
}

func TestClose_ConcurrentCloseAndCredit(t *testing.T) {
	var mu sync.Mutex
	closeCalled := false
	creditCalled := false

	repo := &mocks.MockAccountRepository{
		CloseAtomicallyFunc: func(ctx context.Context, id string) error {
			mu.Lock()
			defer mu.Unlock()
			closeCalled = true
			return nil
		},
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			mu.Lock()
			defer mu.Unlock()
			creditCalled = true
			return domain.ErrAccountClosed
		},
	}
	svc := NewAccountService(repo)

	var wg sync.WaitGroup
	wg.Add(2)

	var closeErr, creditErr error
	go func() {
		defer wg.Done()
		closeErr = svc.Close(context.Background(), "acc-1")
	}()
	go func() {
		defer wg.Done()
		creditErr = svc.Credit(context.Background(), "acc-1", 100)
	}()

	wg.Wait()

	require.NoError(t, closeErr)
	require.ErrorIs(t, creditErr, domain.ErrAccountClosed)

	mu.Lock()
	require.True(t, closeCalled || creditCalled, "at least one operation should have been called")
	mu.Unlock()
}

// ==================== Reserve/Credit/Debit — Edge Cases ====================

func TestReserve_MaxInt64Amount(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, int64(math.MaxInt64), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(context.Background(), "acc-1", math.MaxInt64)

	require.NoError(t, err)
}

func TestReserve_MinInt64Amount(t *testing.T) {
	svc := NewAccountService(&mocks.MockAccountRepository{})

	err := svc.Reserve(context.Background(), "acc-1", math.MinInt64)

	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestCredit_LargeAmount(t *testing.T) {
	const oneBillion int64 = 1_000_000_000
	repo := &mocks.MockAccountRepository{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, oneBillion, amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Credit(context.Background(), "acc-1", oneBillion)

	require.NoError(t, err)
}

func TestDebit_ExactBalance(t *testing.T) {
	const balance int64 = 5000
	repo := &mocks.MockAccountRepository{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			assert.Equal(t, balance, amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Debit(context.Background(), "acc-1", balance)

	require.NoError(t, err)
}

func TestDebit_OnlyReservedFunds(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.Debit(context.Background(), "acc-1", 500)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestReserve_Credit_Debit_Sequence(t *testing.T) {
	var mu sync.Mutex
	var ops []string

	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			mu.Lock()
			ops = append(ops, "reserve")
			mu.Unlock()
			assert.Equal(t, int64(300), amount)
			return nil
		},
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			mu.Lock()
			ops = append(ops, "credit")
			mu.Unlock()
			assert.Equal(t, int64(500), amount)
			return nil
		},
		DebitFunc: func(ctx context.Context, id string, amount int64) error {
			mu.Lock()
			ops = append(ops, "debit")
			mu.Unlock()
			assert.Equal(t, int64(200), amount)
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(context.Background(), "acc-1", 300)
	require.NoError(t, err)

	err = svc.Credit(context.Background(), "acc-1", 500)
	require.NoError(t, err)

	err = svc.Debit(context.Background(), "acc-1", 200)
	require.NoError(t, err)

	mu.Lock()
	require.Equal(t, []string{"reserve", "credit", "debit"}, ops)
	mu.Unlock()
}

// ==================== CommitReservation/CancelReservation — Edge Cases ====================

func TestCommitReservation_AfterCancel(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrInsufficientFunds
		},
	}
	svc := NewAccountService(repo)

	err := svc.CommitReservation(context.Background(), "acc-1", 100)

	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestCancelReservation_AfterCommit(t *testing.T) {
	repo := &mocks.MockAccountRepository{
		CancelReservationFunc: func(ctx context.Context, id string, amount int64) error {
			return domain.ErrAccountNotFound
		},
	}
	svc := NewAccountService(repo)

	err := svc.CancelReservation(context.Background(), "acc-1", 100)

	require.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCommitReservation_DoubleCommit(t *testing.T) {
	callCount := 0
	repo := &mocks.MockAccountRepository{
		CommitReservationFunc: func(ctx context.Context, id string, amount int64) error {
			callCount++
			if callCount > 1 {
				return domain.ErrInsufficientFunds
			}
			return nil
		},
	}
	svc := NewAccountService(repo)

	err := svc.CommitReservation(context.Background(), "acc-1", 200)
	require.NoError(t, err)

	err = svc.CommitReservation(context.Background(), "acc-1", 200)
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

// ==================== Context & Error Propagation ====================

func TestCreate_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	repo := &mocks.MockAccountRepository{
		CreateFunc: func(ctx context.Context, account *domain.Account) error {
			return ctx.Err()
		},
	}
	svc := NewAccountService(repo)

	result, err := svc.Create(ctx, domain.CreateAccountInput{OwnerName: "Timeout"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestReserve_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mocks.MockAccountRepository{
		ReserveFunc: func(ctx context.Context, id string, amount int64) error {
			return ctx.Err()
		},
	}
	svc := NewAccountService(repo)

	err := svc.Reserve(ctx, "acc-1", 100)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCredit_RepoReturnsWrappedError(t *testing.T) {
	wrappedErr := errors.New("connection reset by peer: timeout exceeded")
	repo := &mocks.MockAccountRepository{
		CreditFunc: func(ctx context.Context, id string, amount int64) error {
			return wrappedErr
		},
	}
	svc := NewAccountService(repo)

	err := svc.Credit(context.Background(), "acc-1", 100)

	require.Error(t, err)
	assert.EqualError(t, err, "connection reset by peer: timeout exceeded")
}
