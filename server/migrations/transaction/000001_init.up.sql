CREATE TABLE IF NOT EXISTS transactions (
    id               UUID PRIMARY KEY,
    from_account_id  UUID NOT NULL REFERENCES accounts(id),
    to_account_id    UUID NOT NULL REFERENCES accounts(id),
    amount           BIGINT NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    idempotency_key  VARCHAR(255) NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_idempotency_key ON transactions(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

ALTER TABLE transactions ADD CONSTRAINT chk_transactions_status
    CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'COMPENSATED'));
