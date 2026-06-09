-- +goose Up
CREATE TABLE settings (
    id          SERIAL PRIMARY KEY,
    key         VARCHAR(255) NOT NULL UNIQUE,
    value       JSONB NOT NULL DEFAULT '{}',
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS settings;