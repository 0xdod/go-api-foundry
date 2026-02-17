ALTER TABLE accounts
ADD COLUMN name VARCHAR(255),
ADD COLUMN code CITEXT UNIQUE,
ADD COLUMN currency CITEXT NOT NULL DEFAULT 'NGN',
ADD COLUMN normal_balance CITEXT NOT NULL DEFAULT 'credit',
ADD COLUMN user_id BIGINT NULL;

UPDATE accounts SET normal_balance = 'debit' WHERE type = 'treasury';
UPDATE accounts SET normal_balance = 'debit' WHERE type = 'external_funding';
UPDATE accounts SET normal_balance = 'credit' WHERE type = 'fee';
