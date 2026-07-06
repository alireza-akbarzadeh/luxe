-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS public_collections (
    id              BIGSERIAL PRIMARY KEY,
    slug            TEXT         NOT NULL UNIQUE,
    title           TEXT         NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    theme           TEXT         NOT NULL DEFAULT '',
    cover_image_url TEXT         NOT NULL DEFAULT '',
    author_name     TEXT         NOT NULL DEFAULT '',
    author_handle   TEXT         NOT NULL DEFAULT '',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_public_collections_is_active ON public_collections(is_active);
CREATE INDEX IF NOT EXISTS idx_public_collections_sort_order ON public_collections(sort_order);
CREATE INDEX IF NOT EXISTS idx_public_collections_deleted_at ON public_collections(deleted_at);

CREATE TABLE IF NOT EXISTS public_collection_items (
    id          BIGSERIAL PRIMARY KEY,
    collection_id BIGINT       NOT NULL REFERENCES public_collections(id) ON DELETE CASCADE,
    product_id  BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    note        TEXT         NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_public_collection_items_collection_id ON public_collection_items(collection_id);
CREATE INDEX IF NOT EXISTS idx_public_collection_items_product_id ON public_collection_items(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public_collection_items;
DROP TABLE IF EXISTS public_collections;
-- +goose StatementEnd
