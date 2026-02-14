CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    type CITEXT NOT NULL DEFAULT 'user'
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reference CITEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    transaction_id BIGSERIAL NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    account_id BIGSERIAL NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount BIGINT NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('debit','credit')),
    balance_after BIGINT NOT NULL,
    locked_balance_after BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS balances (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    account_id BIGSERIAL NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0,
    locked_balance BIGINT NOT NULL DEFAULT 0
);

-- seed system accounts
WITH account_types AS (
    SELECT unnest(array['treasury', 'external_funding', 'fee']) AS type
)
INSERT INTO accounts (type) 
SELECT type FROM account_types;


