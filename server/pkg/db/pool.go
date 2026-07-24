package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is an interface abstracting pgxpool.Pool operations used by repositories.
type Pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Compile-time check that *pgxpool.Pool implements Pool.
var _ Pool = (*pgxpool.Pool)(nil)

// ExecAffected executes a query and returns notFoundErr if no rows were affected.
func ExecAffected(ctx context.Context, pool Pool, query string, notFoundErr error, args ...any) error {
	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notFoundErr
	}
	return nil
}
