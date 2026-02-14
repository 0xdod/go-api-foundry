-- name: CreateAccount :one
INSERT INTO accounts (type) VALUES ($1) RETURNING *;

-- name: FindAccountByID :one
SELECT * FROM accounts WHERE id = $1;

-- name: FindAccountByType :one
SELECT * FROM accounts WHERE type = $1;

-- name: ListAccounts :many
SELECT * FROM accounts;

-- name: UpdateAccount :one
UPDATE accounts SET type = $2 WHERE id = $1 RETURNING *;

-- name: DeleteAccount :one
DELETE FROM accounts WHERE id = $1 RETURNING *;

