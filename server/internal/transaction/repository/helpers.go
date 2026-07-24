package repository

import (
	"context"

	"go-core-banking-system/pkg/db"
)

func execAffected(ctx context.Context, pool db.Pool, query string, notFoundErr error, args ...any) error {
	return db.ExecAffected(ctx, pool, query, notFoundErr, args...)
}
