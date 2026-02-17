CREATE TABLE IF NOT EXISTS idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    user_id BIGINT, 
    resource_path TEXT NOT NULL,
    request_params JSONB, 
    response_code INT,
    response_body JSONB,
    recovery_point TEXT DEFAULT 'started', 
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (key, user_id) 
);

CREATE INDEX idx_idempotency_keys_key ON idempotency_keys(key);
