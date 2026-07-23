package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

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
	publisher  domain.EventPublisher
}

func NewTransactionService(repo domain.Repository, accountCli AccountClient, publisher domain.EventPublisher) domain.Service {
	return &transactionService{repo: repo, accountCli: accountCli, publisher: publisher}
}

func (s *transactionService) Transfer(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
	if err := validateTransferInput(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}

	tx := s.createTransaction(input)

	if err := s.repo.Create(ctx, tx); err != nil {
		if isUniqueViolation(err) {
			return s.repo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		}
		return nil, err
	}

	if err := s.executeSaga(ctx, tx); err != nil {
		s.updateStatus(ctx, tx.ID, domain.TxStatusFailed)
		tx.Status = domain.TxStatusFailed
		s.publishEvent(tx, domain.EventTransactionFailed)
		return tx, err
	}

	s.updateStatus(ctx, tx.ID, domain.TxStatusCompleted)
	tx.Status = domain.TxStatusCompleted
	s.publishEvent(tx, domain.EventTransactionCompleted)
	return tx, nil
}

func validateTransferInput(input domain.TransferInput) error {
	if input.FromAccountID == input.ToAccountID {
		return domain.ErrSameAccount
	}
	if input.Amount <= 0 {
		return domain.ErrInvalidAmount
	}
	if input.IdempotencyKey == "" {
		return domain.ErrIdempotencyKeyRequired
	}
	return nil
}

func (s *transactionService) createTransaction(input domain.TransferInput) *domain.Transaction {
	now := time.Now()
	return &domain.Transaction{
		ID:             uuid.New().String(),
		FromAccountID:  input.FromAccountID,
		ToAccountID:    input.ToAccountID,
		Amount:         input.Amount,
		Status:         domain.TxStatusPending,
		IdempotencyKey: input.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *transactionService) updateStatus(ctx context.Context, txID string, status domain.TransactionStatus) {
	if err := s.repo.UpdateStatus(ctx, txID, status); err != nil {
		slog.Error("failed to update transaction status", "tx_id", txID, "error", err)
	}
}

func (s *transactionService) executeSaga(ctx context.Context, tx *domain.Transaction) error {
	if err := s.accountCli.Reserve(ctx, tx.FromAccountID, tx.Amount); err != nil {
		return err
	}

	if err := s.accountCli.Credit(ctx, tx.ToAccountID, tx.Amount); err != nil {
		s.compensateCancelReservation(ctx, tx)
		return err
	}

	if err := s.accountCli.CommitReservation(ctx, tx.FromAccountID, tx.Amount); err != nil {
		s.compensateFull(ctx, tx)
		return err
	}

	return nil
}

func (s *transactionService) compensateCancelReservation(ctx context.Context, tx *domain.Transaction) {
	if err := s.accountCli.CancelReservation(ctx, tx.FromAccountID, tx.Amount); err != nil {
		slog.Error("compensation failed", "tx_id", tx.ID, "step", "cancel_reservation", "error", err)
	}
}

func (s *transactionService) compensateFull(ctx context.Context, tx *domain.Transaction) {
	s.compensateCancelReservation(ctx, tx)
	if err := s.accountCli.Debit(ctx, tx.ToAccountID, tx.Amount); err != nil {
		slog.Error("compensation failed", "tx_id", tx.ID, "step", "debit", "error", err)
	}
}

func (s *transactionService) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (s *transactionService) publishEvent(tx *domain.Transaction, eventType domain.EventType) {
	if s.publisher == nil {
		return
	}

	event := &domain.TransactionEvent{
		TransactionID:  tx.ID,
		Type:           eventType,
		Status:         tx.Status,
		FromAccountID:  tx.FromAccountID,
		ToAccountID:    tx.ToAccountID,
		Amount:         tx.Amount,
		IdempotencyKey: tx.IdempotencyKey,
		Timestamp:      time.Now(),
	}

	_ = s.publisher.PublishTransactionEvent(event)
}

func (s *transactionService) publishEvent(tx *domain.Transaction, eventType domain.EventType) {
	if s.publisher == nil {
		return
	}

	event := &domain.TransactionEvent{
		TransactionID:  tx.ID,
		Type:           eventType,
		Status:         tx.Status,
		FromAccountID:  tx.FromAccountID,
		ToAccountID:    tx.ToAccountID,
		Amount:         tx.Amount,
		IdempotencyKey: tx.IdempotencyKey,
		Timestamp:      time.Now(),
	}

	_ = s.publisher.PublishTransactionEvent(event)
}
