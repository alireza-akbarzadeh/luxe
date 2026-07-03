-- +goose Up
-- +goose StatementBegin
ALTER TABLE collections
    ADD COLUMN IF NOT EXISTS theme TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_collections_theme ON collections(theme);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_collections_theme;
ALTER TABLE collections DROP COLUMN IF EXISTS theme;
-- +goose StatementEnd
