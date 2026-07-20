CREATE TABLE IF NOT EXISTS accounts (
    id          UUID PRIMARY KEY,
    owner_name  VARCHAR(255) NOT NULL,
    balance     BIGINT NOT NULL DEFAULT 0,
    status      VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_owner_name ON accounts(owner_name);

ALTER TABLE accounts ADD CONSTRAINT chk_accounts_status
    CHECK (status IN ('ACTIVE', 'BLOCKED', 'CLOSED'));
