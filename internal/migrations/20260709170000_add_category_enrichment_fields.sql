-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS icon TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS meta_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS meta_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_categories_sort_order;

ALTER TABLE categories
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS meta_description,
    DROP COLUMN IF EXISTS meta_title,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS icon;
-- +goose StatementEnd
