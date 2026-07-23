package repository

import (
	"context"

	"go-core-banking-system/pkg/db"
)

func execAffected(ctx context.Context, pool db.Pool, query string, notFoundErr error, args ...any) error {
	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notFoundErr
	}
	return nil
}
