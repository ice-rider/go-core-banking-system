package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/account/domain"
)

type mockPool struct {
	execFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return nil
}

type mockRow struct {
	scanFunc func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.scanFunc != nil {
		return r.scanFunc(dest...)
	}
	return nil
}

func TestNewPostgresRepo(t *testing.T) {
	pool := &mockPool{}
	repo := NewPostgresRepo(pool)
	require.NotNil(t, repo)
}

func TestCreate_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
	}
	repo := NewPostgresRepo(pool)

	account := &domain.Account{
		ID:        "acc-123",
		OwnerName: "Alice",
		Balance:   1000,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), account)
	assert.NoError(t, err)
}

func TestCreate_Error(t *testing.T) {
	expectedErr := errors.New("connection refused")
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, expectedErr
		},
	}
	repo := NewPostgresRepo(pool)

	account := &domain.Account{
		ID:        "acc-123",
		OwnerName: "Alice",
		Balance:   1000,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), account)
	assert.ErrorIs(t, err, expectedErr)
}

func TestGetByID_Success(t *testing.T) {
	now := time.Now()
	pool := &mockPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...any) error {
					*dest[0].(*string) = "acc-123"
					*dest[1].(*string) = "Alice"
					*dest[2].(*int64) = 1000
					*dest[3].(*domain.Status) = domain.StatusActive
					*dest[4].(*time.Time) = now
					*dest[5].(*time.Time) = now
					return nil
				},
			}
		},
	}
	repo := NewPostgresRepo(pool)

	account, err := repo.GetByID(context.Background(), "acc-123")
	require.NoError(t, err)
	assert.Equal(t, "acc-123", account.ID)
	assert.Equal(t, "Alice", account.OwnerName)
	assert.Equal(t, int64(1000), account.Balance)
	assert.Equal(t, domain.StatusActive, account.Status)
	assert.Equal(t, now, account.CreatedAt)
	assert.Equal(t, now, account.UpdatedAt)
}

func TestGetByID_NotFound(t *testing.T) {
	pool := &mockPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...any) error {
					return pgx.ErrNoRows
				},
			}
		},
	}
	repo := NewPostgresRepo(pool)

	account, err := repo.GetByID(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.Nil(t, account)
}

func TestGetByID_ScanError(t *testing.T) {
	expectedErr := errors.New("scan failed")
	pool := &mockPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...any) error {
					return expectedErr
				},
			}
		},
	}
	repo := NewPostgresRepo(pool)

	account, err := repo.GetByID(context.Background(), "acc-123")
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, account)
}

func TestUpdateStatus_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.UpdateStatus(context.Background(), "acc-123", domain.StatusBlocked)
	assert.NoError(t, err)
}

func TestUpdateStatus_ExecError(t *testing.T) {
	expectedErr := errors.New("exec failed")
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, expectedErr
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.UpdateStatus(context.Background(), "acc-123", domain.StatusBlocked)
	assert.ErrorIs(t, err, expectedErr)
}

func TestReserve_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Reserve(context.Background(), "acc-123", 500)
	assert.NoError(t, err)
}

func TestReserve_InsufficientFunds(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Reserve(context.Background(), "acc-123", 500)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestReserve_ExecError(t *testing.T) {
	expectedErr := errors.New("exec failed")
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, expectedErr
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Reserve(context.Background(), "acc-123", 500)
	assert.ErrorIs(t, err, expectedErr)
}

func TestCredit_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Credit(context.Background(), "acc-123", 500)
	assert.NoError(t, err)
}

func TestCredit_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Credit(context.Background(), "nonexistent", 500)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestDebit_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Debit(context.Background(), "acc-123", 500)
	assert.NoError(t, err)
}

func TestDebit_InsufficientFunds(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.Debit(context.Background(), "acc-123", 500)
	assert.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestCommitReservation_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CommitReservation(context.Background(), "acc-123", 500)
	assert.NoError(t, err)
}

func TestCommitReservation_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CommitReservation(context.Background(), "nonexistent", 500)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCancelReservation_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CancelReservation(context.Background(), "acc-123", 500)
	assert.NoError(t, err)
}

func TestCancelReservation_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CancelReservation(context.Background(), "nonexistent", 500)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestBlockAtomically_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.BlockAtomically(context.Background(), "acc-123")
	assert.NoError(t, err)
}

func TestBlockAtomically_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.BlockAtomically(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestUnblockAtomically_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.UnblockAtomically(context.Background(), "acc-123")
	assert.NoError(t, err)
}

func TestUnblockAtomically_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.UnblockAtomically(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestCloseAtomically_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CloseAtomically(context.Background(), "acc-123")
	assert.NoError(t, err)
}

func TestCloseAtomically_BalanceNotZero(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}
	repo := NewPostgresRepo(pool)

	err := repo.CloseAtomically(context.Background(), "acc-123")
	assert.ErrorIs(t, err, domain.ErrBalanceNotZero)
}
