-- +goose Up
-- +goose StatementBegin

ALTER TABLE collections
    ADD COLUMN IF NOT EXISTS collection_type TEXT NOT NULL DEFAULT 'smart',
    ADD COLUMN IF NOT EXISTS starts_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS ends_at     TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_collections_collection_type ON collections(collection_type);
CREATE INDEX IF NOT EXISTS idx_collections_starts_at ON collections(starts_at);
CREATE INDEX IF NOT EXISTS idx_collections_ends_at ON collections(ends_at);

CREATE TABLE IF NOT EXISTS collection_products (
    id            BIGSERIAL PRIMARY KEY,
    collection_id BIGINT      NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    product_id    BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sort_order    INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (collection_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_collection_products_collection_id ON collection_products(collection_id);
CREATE INDEX IF NOT EXISTS idx_collection_products_product_id ON collection_products(product_id);
CREATE INDEX IF NOT EXISTS idx_collection_products_sort_order ON collection_products(sort_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS collection_products;

DROP INDEX IF EXISTS idx_collections_ends_at;
DROP INDEX IF EXISTS idx_collections_starts_at;
DROP INDEX IF EXISTS idx_collections_collection_type;

ALTER TABLE collections
    DROP COLUMN IF EXISTS ends_at,
    DROP COLUMN IF EXISTS starts_at,
    DROP COLUMN IF EXISTS collection_type;

-- +goose StatementEnd
