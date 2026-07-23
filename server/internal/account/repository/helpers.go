package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func execAffected(ctx context.Context, pool *pgxpool.Pool, query string, notFoundErr error, args ...any) error {
	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notFoundErr
	}
	return nil
}
