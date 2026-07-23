package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-core-banking-system/internal/account/domain"
	"go-core-banking-system/pkg/proto/account"
)

type AccountGRPCHandler struct {
	account.UnimplementedAccountServiceServer
	svc domain.Service
}

func NewAccountGRPCHandler(svc domain.Service) *AccountGRPCHandler {
	return &AccountGRPCHandler{svc: svc}
}

func (h *AccountGRPCHandler) HealthCheck(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "account-service"})
}

func (h *AccountGRPCHandler) Create(ctx context.Context, req *account.CreateAccountRequest) (*account.AccountResponse, error) {
	input := domain.CreateAccountInput{OwnerName: req.OwnerName}
	acc, err := h.svc.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	return toResponse(acc), nil
}

func (h *AccountGRPCHandler) GetByID(ctx context.Context, req *account.GetAccountRequest) (*account.AccountResponse, error) {
	acc, err := h.svc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toResponse(acc), nil
}

func (h *AccountGRPCHandler) Block(ctx context.Context, req *account.BlockAccountRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Block(ctx, req.Id)
}

func (h *AccountGRPCHandler) Unblock(ctx context.Context, req *account.UnblockAccountRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Unblock(ctx, req.Id)
}

func (h *AccountGRPCHandler) Close(ctx context.Context, req *account.CloseAccountRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Close(ctx, req.Id)
}

func (h *AccountGRPCHandler) Reserve(ctx context.Context, req *account.ReserveRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Reserve(ctx, req.Id, req.Amount)
}

func (h *AccountGRPCHandler) Credit(ctx context.Context, req *account.CreditRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Credit(ctx, req.Id, req.Amount)
}

func (h *AccountGRPCHandler) Debit(ctx context.Context, req *account.DebitRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.Debit(ctx, req.Id, req.Amount)
}

func (h *AccountGRPCHandler) CommitReservation(ctx context.Context, req *account.CommitReservationRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.CommitReservation(ctx, req.Id, req.Amount)
}

func (h *AccountGRPCHandler) CancelReservation(ctx context.Context, req *account.CancelReservationRequest) (*account.Empty, error) {
	return &account.Empty{}, h.svc.CancelReservation(ctx, req.Id, req.Amount)
}

func toResponse(acc *domain.Account) *account.AccountResponse {
	return &account.AccountResponse{
		Id:        acc.ID,
		OwnerName: acc.OwnerName,
		Balance:   acc.Balance,
		Status:    string(acc.Status),
		CreatedAt: acc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: acc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
