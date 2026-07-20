package repository

const (
	queryCreateTransaction = `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	queryGetTransactionByID = `
		SELECT id, from_account_id, to_account_id, amount, status, idempotency_key, created_at, updated_at
		FROM transactions WHERE id = $1`

	queryUpdateTransactionStatus = `
		UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`

	queryGetTransactionByIDempotencyKey = `
		SELECT id, from_account_id, to_account_id, amount, status, idempotency_key, created_at, updated_at
		FROM transactions WHERE idempotency_key = $1`
)
