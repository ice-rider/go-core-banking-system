package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryConstant_CreateTransaction(t *testing.T) {
	assert.NotEmpty(t, queryCreateTransaction)
	assert.Contains(t, queryCreateTransaction, "INSERT INTO transactions")
	assert.Contains(t, queryCreateTransaction, "$1")
	assert.Contains(t, queryCreateTransaction, "$2")
	assert.Contains(t, queryCreateTransaction, "$3")
	assert.Contains(t, queryCreateTransaction, "$4")
	assert.Contains(t, queryCreateTransaction, "$5")
	assert.Contains(t, queryCreateTransaction, "$6")
	assert.Contains(t, queryCreateTransaction, "$7")
	assert.Contains(t, queryCreateTransaction, "$8")
}

func TestQueryConstant_GetTransactionByID(t *testing.T) {
	assert.NotEmpty(t, queryGetTransactionByID)
	assert.Contains(t, queryGetTransactionByID, "SELECT")
	assert.Contains(t, queryGetTransactionByID, "FROM transactions")
	assert.Contains(t, queryGetTransactionByID, "WHERE id = $1")
}

func TestQueryConstant_UpdateTransactionStatus(t *testing.T) {
	assert.NotEmpty(t, queryUpdateTransactionStatus)
	assert.Contains(t, queryUpdateTransactionStatus, "UPDATE transactions SET status = $1")
	assert.Contains(t, queryUpdateTransactionStatus, "$2")
}

func TestQueryConstant_GetTransactionByIDempotencyKey(t *testing.T) {
	assert.NotEmpty(t, queryGetTransactionByIDempotencyKey)
	assert.Contains(t, queryGetTransactionByIDempotencyKey, "SELECT")
	assert.Contains(t, queryGetTransactionByIDempotencyKey, "FROM transactions")
	assert.Contains(t, queryGetTransactionByIDempotencyKey, "WHERE idempotency_key = $1")
}

func TestAllQueries_NonEmpty(t *testing.T) {
	queries := map[string]string{
		"queryCreateTransaction":              queryCreateTransaction,
		"queryGetTransactionByID":             queryGetTransactionByID,
		"queryUpdateTransactionStatus":        queryUpdateTransactionStatus,
		"queryGetTransactionByIDempotencyKey": queryGetTransactionByIDempotencyKey,
	}

	for name, q := range queries {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, q)
			assert.True(t, strings.TrimSpace(q) != "")
		})
	}
}

func TestQueryParameterOrdering(t *testing.T) {
	tests := []struct {
		name  string
		query string
		count int
	}{
		{
			name:  "CreateTransaction uses 8 parameters",
			query: queryCreateTransaction,
			count: 8,
		},
		{
			name:  "GetTransactionByID uses 1 parameter",
			query: queryGetTransactionByID,
			count: 1,
		},
		{
			name:  "UpdateTransactionStatus uses 2 parameters",
			query: queryUpdateTransactionStatus,
			count: 2,
		},
		{
			name:  "GetTransactionByIDempotencyKey uses 1 parameter",
			query: queryGetTransactionByIDempotencyKey,
			count: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.count > 0)
			for i := 1; i <= tt.count; i++ {
				param := "$" + string(rune('0'+i))
				assert.Contains(t, tt.query, param, "query should contain parameter %s", param)
			}
		})
	}
}
