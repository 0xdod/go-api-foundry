-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (
    key, 
    user_id, 
    resource_path, 
    request_params, 
    recovery_point
) VALUES (
    $1, $2, $3, $4, 'started'
) RETURNING *;

-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys 
WHERE key = $1 AND user_id = $2;

-- name: UpdateIdempotencyKeyResponse :one
UPDATE idempotency_keys 
SET 
    response_code = $3, 
    response_body = $4, 
    message = $5,
    recovery_point = 'completed'
WHERE key = $1 AND user_id = $2
RETURNING *;

-- name: DeleteIdempotencyKey :exec
DELETE FROM idempotency_keys 
WHERE key = $1 AND user_id = $2;

