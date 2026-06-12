-- +goose Up
-- +goose StatementBegin
CREATE TABLE product_attributes (
                                    id          BIGSERIAL PRIMARY KEY,
                                    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                                    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                                    deleted_at  TIMESTAMPTZ,
                                    product_id  BIGINT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
                                    name        TEXT NOT NULL,
                                    values      TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_product_attributes_product_id ON product_attributes (product_id);
CREATE INDEX idx_product_attributes_deleted_at ON product_attributes (deleted_at);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE products
    ADD COLUMN track_inventory     BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN warehouse_location  TEXT,
    ADD COLUMN allow_backorder     BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN visibility          TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    ADD COLUMN tags                TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN channels            TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN published_at        TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
    DROP COLUMN IF EXISTS track_inventory,
    DROP COLUMN IF EXISTS warehouse_location,
    DROP COLUMN IF EXISTS allow_backorder,
    DROP COLUMN IF EXISTS visibility,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS channels,
    DROP COLUMN IF EXISTS published_at;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS product_attributes;
-- +goose StatementEnd