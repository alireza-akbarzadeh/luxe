-- +goose Up
-- +goose StatementBegin
ALTER TABLE returns
    ADD COLUMN IF NOT EXISTS return_type VARCHAR(32) NOT NULL DEFAULT 'refund',
    ADD COLUMN IF NOT EXISTS admin_notes TEXT,
    ADD COLUMN IF NOT EXISTS exchange_notes TEXT;

CREATE INDEX IF NOT EXISTS idx_returns_return_type ON returns (return_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_returns_return_type;
ALTER TABLE returns
    DROP COLUMN IF EXISTS exchange_notes,
    DROP COLUMN IF EXISTS admin_notes,
    DROP COLUMN IF EXISTS return_type;
-- +goose StatementEnd
