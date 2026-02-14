-- name: SaveBalance :one
INSERT INTO balances (account_id, balance, locked_balance)
VALUES ($1, $2, $3)
ON CONFLICT (account_id) DO UPDATE
SET balance = COALESCE(sqlc.narg('balance')::bigint, balances.balance),
    locked_balance = COALESCE(sqlc.narg('locked_balance')::bigint, locked_balance)
RETURNING *;

-- name: AddToBalance :one
INSERT INTO balances (account_id, balance, locked_balance)
VALUES ($1, $2, $3)
ON CONFLICT (account_id) DO UPDATE
SET balance = COALESCE(balances.balance, 0) + COALESCE(sqlc.narg('balance')::bigint, 0),
    locked_balance = COALESCE(balances.locked_balance, 0) + COALESCE(sqlc.narg('locked_balance')::bigint, 0)
RETURNING *;

-- name: GetBalance :one
SELECT * FROM balances WHERE account_id = $1;

-- name: GetBalances :many
SELECT * FROM balances;

-- name: DeleteBalance :one
DELETE FROM balances WHERE account_id = $1 RETURNING *;