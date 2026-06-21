-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Products: localized fields + cross-locale search document
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS name_i18n JSONB,
    ADD COLUMN IF NOT EXISTS description_i18n JSONB,
    ADD COLUMN IF NOT EXISTS search_aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS search_document TEXT;

UPDATE products
SET name_i18n = jsonb_build_object('en', name)
WHERE name_i18n IS NULL AND name <> '';

UPDATE products
SET description_i18n = jsonb_build_object('en', description)
WHERE description_i18n IS NULL AND description IS NOT NULL AND description <> '';

UPDATE products
SET search_document = trim(both FROM concat_ws(' ', name, description, sku, barcode))
WHERE search_document IS NULL OR search_document = '';

DROP INDEX IF EXISTS idx_products_search;
ALTER TABLE products DROP COLUMN IF EXISTS search_vector;

ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', coalesce(search_document, ''))) STORED;

CREATE INDEX idx_products_search ON products USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_products_search_document_trgm ON products USING GIN (search_document gin_trgm_ops);

-- Categories: localized fields + search document
ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS name_i18n JSONB,
    ADD COLUMN IF NOT EXISTS description_i18n JSONB,
    ADD COLUMN IF NOT EXISTS search_document TEXT;

UPDATE categories
SET name_i18n = jsonb_build_object('en', name)
WHERE name_i18n IS NULL AND name <> '';

UPDATE categories
SET description_i18n = jsonb_build_object('en', description)
WHERE description_i18n IS NULL AND description IS NOT NULL AND description <> '';

UPDATE categories
SET search_document = trim(both FROM concat_ws(' ', name, description))
WHERE search_document IS NULL OR search_document = '';

DROP INDEX IF EXISTS idx_categories_search;
ALTER TABLE categories DROP COLUMN IF EXISTS search_vector;

ALTER TABLE categories ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', coalesce(search_document, ''))) STORED;

CREATE INDEX idx_categories_search ON categories USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_categories_search_document_trgm ON categories USING GIN (search_document gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_products_search_document_trgm;
DROP INDEX IF EXISTS idx_categories_search_document_trgm;
DROP INDEX IF EXISTS idx_products_search;
DROP INDEX IF EXISTS idx_categories_search;

ALTER TABLE products DROP COLUMN IF EXISTS search_vector;
ALTER TABLE categories DROP COLUMN IF EXISTS search_vector;

ALTER TABLE products ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'B')
    ) STORED;

ALTER TABLE categories ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(name, '')), 'A')
    ) STORED;

CREATE INDEX idx_products_search ON products USING GIN (search_vector);
CREATE INDEX idx_categories_search ON categories USING GIN (search_vector);

ALTER TABLE products
    DROP COLUMN IF EXISTS search_aliases,
    DROP COLUMN IF EXISTS search_document,
    DROP COLUMN IF EXISTS description_i18n,
    DROP COLUMN IF EXISTS name_i18n;

ALTER TABLE categories
    DROP COLUMN IF EXISTS search_document,
    DROP COLUMN IF EXISTS description_i18n,
    DROP COLUMN IF EXISTS name_i18n;

-- +goose StatementEnd
