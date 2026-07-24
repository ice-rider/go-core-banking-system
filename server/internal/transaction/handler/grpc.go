package handler

import (
	"context"

	"go-core-banking-system/internal/transaction/domain"
	"go-core-banking-system/pkg/proto/transaction"
)

type TransactionGRPCHandler struct {
	transaction.UnimplementedTransactionServiceServer
	svc domain.Service
}

func NewTransactionGRPCHandler(svc domain.Service) *TransactionGRPCHandler {
	return &TransactionGRPCHandler{svc: svc}
}

func (h *TransactionGRPCHandler) Transfer(ctx context.Context, req *transaction.TransferRequest) (*transaction.TransactionResponse, error) {
	input := domain.TransferInput{
		FromAccountID:  req.FromAccountId,
		ToAccountID:    req.ToAccountId,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	}
	tx, err := h.svc.Transfer(ctx, input)
	if err != nil {
		return nil, err
	}
	return toResponse(tx), nil
}

func (h *TransactionGRPCHandler) GetByID(ctx context.Context, req *transaction.GetTransactionRequest) (*transaction.TransactionResponse, error) {
	tx, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toResponse(tx), nil
}

func toResponse(tx *domain.Transaction) *transaction.TransactionResponse {
	return &transaction.TransactionResponse{
		Id:             tx.ID,
		FromAccountId:  tx.FromAccountID,
		ToAccountId:    tx.ToAccountID,
		Amount:         tx.Amount,
		Status:         string(tx.Status),
		IdempotencyKey: tx.IdempotencyKey,
		CreatedAt:      tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      tx.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
