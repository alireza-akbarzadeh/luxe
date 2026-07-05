-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS community_shopping_lists (
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

CREATE INDEX IF NOT EXISTS idx_community_shopping_lists_is_active ON community_shopping_lists(is_active);
CREATE INDEX IF NOT EXISTS idx_community_shopping_lists_sort_order ON community_shopping_lists(sort_order);
CREATE INDEX IF NOT EXISTS idx_community_shopping_lists_deleted_at ON community_shopping_lists(deleted_at);

CREATE TABLE IF NOT EXISTS community_shopping_list_items (
    id          BIGSERIAL PRIMARY KEY,
    list_id     BIGINT       NOT NULL REFERENCES community_shopping_lists(id) ON DELETE CASCADE,
    product_id  BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    note        TEXT         NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_community_shopping_list_items_list_id ON community_shopping_list_items(list_id);
CREATE INDEX IF NOT EXISTS idx_community_shopping_list_items_product_id ON community_shopping_list_items(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS community_shopping_list_items;
DROP TABLE IF EXISTS community_shopping_lists;
-- +goose StatementEnd
