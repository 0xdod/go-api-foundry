-- name: CreateAccount :one
INSERT INTO accounts (name, code, currency, user_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: FindAccountByID :one
SELECT a.*, COALESCE(b.balance, 0) as balance, COALESCE(b.locked_balance, 0) as locked_balance 
FROM accounts a LEFT JOIN balances b ON b.account_id = a.id 
WHERE a.id = $1;

-- name: FindAccountByType :one
SELECT a.*, COALESCE(b.balance, 0) as balance, COALESCE(b.locked_balance, 0) as locked_balance 
FROM accounts a LEFT JOIN balances b ON b.account_id = a.id 
WHERE a.type = $1;

-- name: ListAccounts :many
SELECT a.*, COALESCE(b.balance, 0) as balance, COALESCE(b.locked_balance, 0) as locked_balance 
FROM accounts a LEFT JOIN balances b ON b.account_id = a.id;

-- name: UpdateAccount :one
UPDATE accounts SET name = $2, code = $3, currency = $4, user_id = $5 WHERE id = $1 RETURNING *;

-- name: DeleteAccount :one
DELETE FROM accounts WHERE id = $1 RETURNING *;

-- name: ComputeBalance :one
WITH computed_balance AS (
SELECT
    COALESCE(SUM(
        CASE 
            WHEN a.normal_balance = 'debit' AND e.direction = 'debit'  THEN e.amount
            WHEN a.normal_balance = 'debit' AND e.direction = 'credit' THEN -e.amount
            WHEN a.normal_balance = 'credit' AND e.direction = 'credit' THEN e.amount
            WHEN a.normal_balance = 'credit' AND e.direction = 'debit'  THEN -e.amount
        END
    ), 0) AS balance
FROM accounts a
LEFT JOIN ledger_entries e ON e.account_id = a.id
WHERE a.id = $1
GROUP BY a.id
) 
INSERT INTO balances (account_id, balance) 
SELECT $1, cb.balance
FROM computed_balance cb
ON CONFLICT (account_id) 
    DO UPDATE SET balance = EXCLUDED.balance
RETURNING *;


