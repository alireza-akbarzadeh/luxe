-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS membership_tier VARCHAR(32) NOT NULL DEFAULT 'free',
    ADD COLUMN IF NOT EXISTS plus_subscribed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS plus_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_membership_tier ON users (membership_tier);

-- +goose Down
DROP INDEX IF EXISTS idx_users_membership_tier;
ALTER TABLE users
    DROP COLUMN IF EXISTS plus_expires_at,
    DROP COLUMN IF EXISTS plus_subscribed_at,
    DROP COLUMN IF EXISTS membership_tier;
