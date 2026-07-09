-- +goose Up
-- +goose StatementBegin
ALTER TABLE brands
    ADD COLUMN IF NOT EXISTS is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS featured_sort_order INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS meta_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS meta_description TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_brands_is_featured ON brands(is_featured);
CREATE INDEX IF NOT EXISTS idx_brands_featured_sort_order ON brands(featured_sort_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_brands_featured_sort_order;
DROP INDEX IF EXISTS idx_brands_is_featured;

ALTER TABLE brands
    DROP COLUMN IF EXISTS meta_description,
    DROP COLUMN IF EXISTS meta_title,
    DROP COLUMN IF EXISTS featured_sort_order,
    DROP COLUMN IF EXISTS is_featured;
-- +goose StatementEnd
