package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-core-banking-system/internal/transaction/domain"
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
	t.Run("creates repository with nil pool", func(t *testing.T) {
		repo := NewPostgresRepo(nil)
		assert.NotNil(t, repo)
	})

	t.Run("creates repository with mock pool", func(t *testing.T) {
		pool := &mockPool{}
		repo := NewPostgresRepo(pool)
		assert.NotNil(t, repo)
	})
}

func TestCreate_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			assert.Contains(t, sql, "INSERT INTO transactions")
			require.Len(t, arguments, 8)
			return pgconn.CommandTag{}, nil
		},
	}

	repo := NewPostgresRepo(pool)
	now := time.Now()
	tx := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-2",
		Amount:         1000,
		Status:         domain.TxStatusPending,
		IdempotencyKey: "idem-1",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(context.Background(), tx)
	assert.NoError(t, err)
}

func TestCreate_Error(t *testing.T) {
	expectedErr := fmt.Errorf("connection refused")
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, expectedErr
		},
	}

	repo := NewPostgresRepo(pool)
	tx := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-2",
		Amount:         1000,
		Status:         domain.TxStatusPending,
		IdempotencyKey: "idem-1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.Create(context.Background(), tx)
	assert.ErrorIs(t, err, expectedErr)
}

func TestGetByID_Success(t *testing.T) {
	now := time.Now()
	expected := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-2",
		Amount:         1000,
		Status:         domain.TxStatusPending,
		IdempotencyKey: "idem-1",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	pool := &mockPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			assert.Contains(t, sql, "SELECT")
			assert.Contains(t, sql, "FROM transactions")
			require.Len(t, args, 1)
			assert.Equal(t, "tx-1", args[0])
			return &mockRow{
				scanFunc: func(dest ...any) error {
					require.Len(t, dest, 8)
					d := dest
					*(d[0].(*string)) = expected.ID
					*(d[1].(*string)) = expected.FromAccountID
					*(d[2].(*string)) = expected.ToAccountID
					*(d[3].(*int64)) = expected.Amount
					*(d[4].(*domain.TransactionStatus)) = expected.Status
					*(d[5].(*string)) = expected.IdempotencyKey
					*(d[6].(*time.Time)) = expected.CreatedAt
					*(d[7].(*time.Time)) = expected.UpdatedAt
					return nil
				},
			}
		},
	}

	repo := NewPostgresRepo(pool)
	result, err := repo.GetByID(context.Background(), "tx-1")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
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
	result, err := repo.GetByID(context.Background(), "nonexistent")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestGetByID_ScanError(t *testing.T) {
	expectedErr := fmt.Errorf("unexpected scan error")
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
	result, err := repo.GetByID(context.Background(), "tx-1")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, expectedErr)
}

func TestUpdateStatus_Success(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			assert.Contains(t, sql, "UPDATE transactions SET status")
			require.Len(t, arguments, 2)
			assert.Equal(t, domain.TxStatusCompleted, arguments[0])
			assert.Equal(t, "tx-1", arguments[1])
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := NewPostgresRepo(pool)
	err := repo.UpdateStatus(context.Background(), "tx-1", domain.TxStatusCompleted)
	assert.NoError(t, err)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		},
	}

	repo := NewPostgresRepo(pool)
	err := repo.UpdateStatus(context.Background(), "nonexistent", domain.TxStatusCompleted)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestUpdateStatus_ExecError(t *testing.T) {
	expectedErr := fmt.Errorf("deadlock detected")
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, expectedErr
		},
	}

	repo := NewPostgresRepo(pool)
	err := repo.UpdateStatus(context.Background(), "tx-1", domain.TxStatusCompleted)
	assert.ErrorIs(t, err, expectedErr)
}

func TestGetByIdempotencyKey_Success(t *testing.T) {
	now := time.Now()
	expected := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-2",
		Amount:         500,
		Status:         domain.TxStatusCompleted,
		IdempotencyKey: "idem-1",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	pool := &mockPool{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			assert.Contains(t, sql, "SELECT")
			assert.Contains(t, sql, "FROM transactions")
			assert.Contains(t, sql, "idempotency_key")
			require.Len(t, args, 1)
			assert.Equal(t, "idem-1", args[0])
			return &mockRow{
				scanFunc: func(dest ...any) error {
					require.Len(t, dest, 8)
					d := dest
					*(d[0].(*string)) = expected.ID
					*(d[1].(*string)) = expected.FromAccountID
					*(d[2].(*string)) = expected.ToAccountID
					*(d[3].(*int64)) = expected.Amount
					*(d[4].(*domain.TransactionStatus)) = expected.Status
					*(d[5].(*string)) = expected.IdempotencyKey
					*(d[6].(*time.Time)) = expected.CreatedAt
					*(d[7].(*time.Time)) = expected.UpdatedAt
					return nil
				},
			}
		},
	}

	repo := NewPostgresRepo(pool)
	result, err := repo.GetByIdempotencyKey(context.Background(), "idem-1")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetByIdempotencyKey_NotFound(t *testing.T) {
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
	result, err := repo.GetByIdempotencyKey(context.Background(), "nonexistent")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestGetByIdempotencyKey_ScanError(t *testing.T) {
	expectedErr := fmt.Errorf("unexpected scan error")
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
	result, err := repo.GetByIdempotencyKey(context.Background(), "idem-1")
	assert.Nil(t, result)
	assert.ErrorIs(t, err, expectedErr)
}

func TestCreate_ArgumentCount(t *testing.T) {
	var capturedArgs []any
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			capturedArgs = arguments
			return pgconn.CommandTag{}, nil
		},
	}

	repo := NewPostgresRepo(pool)
	tx := &domain.Transaction{
		ID:             "tx-1",
		FromAccountID:  "acc-1",
		ToAccountID:    "acc-2",
		Amount:         1000,
		Status:         domain.TxStatusPending,
		IdempotencyKey: "idem-1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := repo.Create(context.Background(), tx)
	require.NoError(t, err)
	require.Len(t, capturedArgs, 8)
	assert.Equal(t, tx.ID, capturedArgs[0])
	assert.Equal(t, tx.FromAccountID, capturedArgs[1])
	assert.Equal(t, tx.ToAccountID, capturedArgs[2])
	assert.Equal(t, tx.Amount, capturedArgs[3])
	assert.Equal(t, tx.Status, capturedArgs[4])
	assert.Equal(t, tx.IdempotencyKey, capturedArgs[5])
	assert.Equal(t, tx.CreatedAt, capturedArgs[6])
	assert.Equal(t, tx.UpdatedAt, capturedArgs[7])
}

func TestUpdateStatus_ArgumentOrder(t *testing.T) {
	var capturedArgs []any
	pool := &mockPool{
		execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
			capturedArgs = arguments
			return pgconn.NewCommandTag("UPDATE 1"), nil
		},
	}

	repo := NewPostgresRepo(pool)
	err := repo.UpdateStatus(context.Background(), "tx-1", domain.TxStatusFailed)
	require.NoError(t, err)
	require.Len(t, capturedArgs, 2)
	assert.Equal(t, domain.TxStatusFailed, capturedArgs[0])
	assert.Equal(t, "tx-1", capturedArgs[1])
}
