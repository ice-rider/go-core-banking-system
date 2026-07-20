package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go-core-banking-system/internal/account/domain"
)

type postgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) domain.Repository {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) Create(ctx context.Context, account *domain.Account) error {
	_, err := r.pool.Exec(ctx, queryCreateAccount,
		account.ID, account.OwnerName, account.Balance,
		account.Status, account.CreatedAt, account.UpdatedAt)
	return err
}

func (r *postgresRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	row := r.pool.QueryRow(ctx, queryGetAccountByID, id)

	var a domain.Account
	err := row.Scan(&a.ID, &a.OwnerName, &a.Balance, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *postgresRepo) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	_, err := r.pool.Exec(ctx, queryUpdateAccountStatus, status, id)
	return err
}

func (r *postgresRepo) Reserve(ctx context.Context, id string, amount int64) error {
	tag, err := r.pool.Exec(ctx, queryReserve, amount, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientFunds
	}
	return nil
}

func (r *postgresRepo) Credit(ctx context.Context, id string, amount int64) error {
	tag, err := r.pool.Exec(ctx, queryCredit, amount, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *postgresRepo) Debit(ctx context.Context, id string, amount int64) error {
	tag, err := r.pool.Exec(ctx, queryDebit, amount, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInsufficientFunds
	}
	return nil
}

// CommitReservation подтверждает резервирование. В нашей модели Reserve уже списал
// средства с баланса, поэтому этот метод лишь фиксирует факт подтверждения (аудит).
func (r *postgresRepo) CommitReservation(ctx context.Context, id string, amount int64) error {
	tag, err := r.pool.Exec(ctx, queryCommitReservation, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *postgresRepo) CancelReservation(ctx context.Context, id string, amount int64) error {
	tag, err := r.pool.Exec(ctx, queryCancelReservation, amount, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *postgresRepo) BlockAtomically(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryBlockAtomically, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *postgresRepo) UnblockAtomically(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryUnblockAtomically, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *postgresRepo) CloseAtomically(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, queryCloseAtomically, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBalanceNotZero
	}
	return nil
}
