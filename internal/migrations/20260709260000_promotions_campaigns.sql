-- +goose Up
ALTER TABLE flash_deals
    ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS starts_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS campaigns (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(128) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    placements JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_campaigns_status_schedule ON campaigns(status, starts_at, ends_at);

-- +goose Down
DROP TABLE IF EXISTS campaigns;
ALTER TABLE flash_deals
    DROP COLUMN IF EXISTS starts_at,
    DROP COLUMN IF EXISTS title;
