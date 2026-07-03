-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS shop_looks (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT         NOT NULL UNIQUE,
    title       TEXT         NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    image_url   TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shop_looks_is_active ON shop_looks(is_active);
CREATE INDEX IF NOT EXISTS idx_shop_looks_sort_order ON shop_looks(sort_order);
CREATE INDEX IF NOT EXISTS idx_shop_looks_deleted_at ON shop_looks(deleted_at);

CREATE TABLE IF NOT EXISTS shop_look_tags (
    id           BIGSERIAL PRIMARY KEY,
    shop_look_id BIGINT       NOT NULL REFERENCES shop_looks(id) ON DELETE CASCADE,
    product_id   BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    x_percent    NUMERIC(5,2) NOT NULL CHECK (x_percent >= 0 AND x_percent <= 100),
    y_percent    NUMERIC(5,2) NOT NULL CHECK (y_percent >= 0 AND y_percent <= 100),
    label        TEXT         NOT NULL DEFAULT '',
    sort_order   INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_look_tags_look_id ON shop_look_tags(shop_look_id);
CREATE INDEX IF NOT EXISTS idx_shop_look_tags_product_id ON shop_look_tags(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shop_look_tags;
DROP TABLE IF EXISTS shop_looks;
-- +goose StatementEnd
