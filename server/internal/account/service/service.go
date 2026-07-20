package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"go-core-banking-system/internal/account/domain"
)

type accountService struct {
	repo domain.Repository
}

func NewAccountService(repo domain.Repository) domain.Service {
	return &accountService{repo: repo}
}

func (s *accountService) Create(ctx context.Context, input domain.CreateAccountInput) (*domain.Account, error) {
	if input.OwnerName == "" {
		return nil, domain.ErrOwnerNameRequired
	}

	now := time.Now()
	account := &domain.Account{
		ID:        uuid.New().String(),
		OwnerName: input.OwnerName,
		Balance:   0,
		Status:    domain.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *accountService) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrAccountNotFound
	}
	return account, nil
}

func (s *accountService) Block(ctx context.Context, id string) error {
	return s.repo.BlockAtomically(ctx, id)
}

func (s *accountService) Unblock(ctx context.Context, id string) error {
	return s.repo.UnblockAtomically(ctx, id)
}

func (s *accountService) Close(ctx context.Context, id string) error {
	return s.repo.CloseAtomically(ctx, id)
}

func (s *accountService) Reserve(ctx context.Context, id string, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	return s.repo.Reserve(ctx, id, amount)
}

func (s *accountService) Credit(ctx context.Context, id string, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	return s.repo.Credit(ctx, id, amount)
}

func (s *accountService) Debit(ctx context.Context, id string, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	return s.repo.Debit(ctx, id, amount)
}

func (s *accountService) CommitReservation(ctx context.Context, id string, amount int64) error {
	return s.repo.CommitReservation(ctx, id, amount)
}

func (s *accountService) CancelReservation(ctx context.Context, id string, amount int64) error {
	return s.repo.CancelReservation(ctx, id, amount)
}
