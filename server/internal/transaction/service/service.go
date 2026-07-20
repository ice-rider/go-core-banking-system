package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-core-banking-system/internal/transaction/domain"
)

type AccountClient interface {
	Reserve(ctx context.Context, id string, amount int64) error
	Credit(ctx context.Context, id string, amount int64) error
	Debit(ctx context.Context, id string, amount int64) error
	CommitReservation(ctx context.Context, id string, amount int64) error
	CancelReservation(ctx context.Context, id string, amount int64) error
}

type transactionService struct {
	repo       domain.Repository
	accountCli AccountClient
}

func NewTransactionService(repo domain.Repository, accountCli AccountClient) domain.Service {
	return &transactionService{repo: repo, accountCli: accountCli}
}

func (s *transactionService) Transfer(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
	if input.FromAccountID == input.ToAccountID {
		return nil, domain.ErrSameAccount
	}
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if input.IdempotencyKey == "" {
		return nil, domain.ErrIdempotencyKeyRequired
	}

	existing, err := s.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}

	now := time.Now()
	tx := &domain.Transaction{
		ID:             uuid.New().String(),
		FromAccountID:  input.FromAccountID,
		ToAccountID:    input.ToAccountID,
		Amount:         input.Amount,
		Status:         domain.TxStatusPending,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		if isUniqueViolation(err) {
			return s.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		}
		return nil, err
	}

	if err := s.executeSaga(ctx, tx); err != nil {
		_ = s.repo.UpdateStatus(ctx, tx.ID, domain.TxStatusFailed)
		return tx, err
	}

	_ = s.repo.UpdateStatus(ctx, tx.ID, domain.TxStatusCompleted)
	tx.Status = domain.TxStatusCompleted
	return tx, nil
}

func (s *transactionService) executeSaga(ctx context.Context, tx *domain.Transaction) error {
	if err := s.accountCli.Reserve(ctx, tx.FromAccountID, tx.Amount); err != nil {
		return err
	}

	if err := s.accountCli.Credit(ctx, tx.ToAccountID, tx.Amount); err != nil {
		_ = s.accountCli.CancelReservation(ctx, tx.FromAccountID, tx.Amount)
		return err
	}

	if err := s.accountCli.CommitReservation(ctx, tx.FromAccountID, tx.Amount); err != nil {
		_ = s.accountCli.CancelReservation(ctx, tx.FromAccountID, tx.Amount)
		_ = s.accountCli.Debit(ctx, tx.ToAccountID, tx.Amount)
		return err
	}

	return nil
}

func (s *transactionService) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unique_violation") || strings.Contains(msg, "duplicate key")
}
