-- name: CreateLedgerEntry :one
INSERT INTO ledger_entries (transaction_id, account_id, amount, direction, balance_after, locked_balance_after)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: FindLedgerEntryByID :one
SELECT * FROM ledger_entries WHERE id = $1;

-- name: FindLedgerEntriesByTransactionID :many
SELECT * FROM ledger_entries WHERE transaction_id = $1;

-- name: FindLedgerEntriesByAccountID :many
SELECT le.*, t.reference, t.type, t.status FROM ledger_entries le
JOIN transactions t ON le.transaction_id = t.id
WHERE account_id = $1
ORDER BY t.created_at DESC;

-- name: ListLedgerEntries :many
SELECT * FROM ledger_entries ORDER BY created_at DESC;

-- name: DeleteLedgerEntry :one
DELETE FROM ledger_entries WHERE id = $1 RETURNING *;