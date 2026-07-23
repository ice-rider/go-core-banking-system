package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-core-banking-system/internal/transaction/domain"
)

type postgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) domain.Repository {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	_, err := r.pool.Exec(ctx, queryCreateTransaction,
		tx.ID, tx.FromAccountID, tx.ToAccountID, tx.Amount,
		tx.Status, tx.IdempotencyKey, tx.CreatedAt, tx.UpdatedAt)
	return err
}

func (r *postgresRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	row := r.pool.QueryRow(ctx, queryGetTransactionByID, id)

	var tx domain.Transaction
	err := row.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount,
		&tx.Status, &tx.IdempotencyKey, &tx.CreatedAt, &tx.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *postgresRepo) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	return execAffected(ctx, r.pool, queryUpdateTransactionStatus, domain.ErrTransactionNotFound, status, id)
}

func (r *postgresRepo) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	row := r.pool.QueryRow(ctx, queryGetTransactionByIDempotencyKey, key)

	var tx domain.Transaction
	err := row.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount,
		&tx.Status, &tx.IdempotencyKey, &tx.CreatedAt, &tx.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
