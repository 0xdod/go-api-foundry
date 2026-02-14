-- name: CreateTransaction :one
INSERT INTO transactions (reference) VALUES ($1) RETURNING *;

-- name: FindTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: FindTransactionByReference :one
SELECT * FROM transactions WHERE reference = $1;

-- name: ListTransactions :many
SELECT * FROM transactions;

-- name: UpdateTransaction :one
UPDATE transactions SET reference = $2 WHERE id = $1 RETURNING *;

-- name: DeleteTransaction :one
DELETE FROM transactions WHERE id = $1 RETURNING *;