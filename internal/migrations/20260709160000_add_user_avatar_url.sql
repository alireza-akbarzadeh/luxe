-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_users_avatar_url ON users (avatar_url)
    WHERE avatar_url <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_users_avatar_url;

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url;
