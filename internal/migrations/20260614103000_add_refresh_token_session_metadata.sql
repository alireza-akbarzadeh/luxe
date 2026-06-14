-- +goose Up
ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS user_agent VARCHAR(512),
ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45),
ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_tokens_last_used_at;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS last_used_at,
DROP COLUMN IF EXISTS ip_address,
DROP COLUMN IF EXISTS user_agent;
