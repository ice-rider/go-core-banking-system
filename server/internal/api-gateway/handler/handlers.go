package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gw_errors "go-core-banking-system/internal/api-gateway/errors"
	"go-core-banking-system/internal/api-gateway/validator"
)

type CreateAccountRequest struct {
	OwnerName string `json:"owner_name" validate:"required,min=1,max=255"`
}

type CreateTransferRequest struct {
	FromAccountID  string `json:"from_account_id" validate:"required,uuid"`
	ToAccountID    string `json:"to_account_id" validate:"required,uuid"`
	Amount         int64  `json:"amount" validate:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" validate:"required,uuid"`
}

type Handler struct {
	accountClient     AccountClient
	transactionClient TransactionClient
	validator         *validator.Validator
}

func NewHandler(ac AccountClient, tc TransactionClient, v *validator.Validator) *Handler {
	return &Handler{accountClient: ac, transactionClient: tc, validator: v}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "invalid request body", Code: "VALIDATION_ERROR"})
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: errs[0].Message, Code: "VALIDATION_ERROR"})
		return
	}

	resp, err := h.accountClient.Create(c.Request.Context(), req.OwnerName)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "id is required", Code: "VALIDATION_ERROR"})
		return
	}

	resp, err := h.accountClient.GetByID(c.Request.Context(), id)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) BlockAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "id is required", Code: "VALIDATION_ERROR"})
		return
	}

	_, err := h.accountClient.Block(c.Request.Context(), id)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account blocked"})
}

func (h *Handler) UnblockAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "id is required", Code: "VALIDATION_ERROR"})
		return
	}

	_, err := h.accountClient.Unblock(c.Request.Context(), id)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account unblocked"})
}

func (h *Handler) CloseAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "id is required", Code: "VALIDATION_ERROR"})
		return
	}

	_, err := h.accountClient.Close(c.Request.Context(), id)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account closed"})
}

func (h *Handler) CreateTransfer(c *gin.Context) {
	var req CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "invalid request body", Code: "VALIDATION_ERROR"})
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: errs[0].Message, Code: "VALIDATION_ERROR"})
		return
	}

	resp, err := h.transactionClient.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.IdempotencyKey)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gw_errors.ErrorResponse{Error: "id is required", Code: "VALIDATION_ERROR"})
		return
	}

	resp, err := h.transactionClient.GetByID(c.Request.Context(), id)
	if err != nil {
		status, errResp := gw_errors.MapError(err)
		c.JSON(status, errResp)
		return
	}

	c.JSON(http.StatusOK, resp)
}
