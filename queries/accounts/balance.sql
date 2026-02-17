-- name: SaveBalance :one
INSERT INTO balances (account_id, balance, locked_balance)
VALUES ($1, $2, $3)
ON CONFLICT (account_id) DO UPDATE
SET balance = COALESCE(EXCLUDED.balance, balances.balance),
    locked_balance = COALESCE(EXCLUDED.locked_balance, balances.locked_balance)
RETURNING *;

-- name: GetBalance :one
SELECT * FROM balances WHERE account_id = $1;

-- name: GetBalances :many
SELECT * FROM balances;

-- name: DeleteBalance :one
DELETE FROM balances WHERE account_id = $1 RETURNING *;