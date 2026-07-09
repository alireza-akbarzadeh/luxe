-- +goose Up
-- +goose StatementBegin

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS parent_order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_orders_parent_order_id ON orders(parent_order_id);

CREATE TABLE IF NOT EXISTS order_tags (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT      NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    tag        TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_order_tags_order_id ON order_tags(order_id);
CREATE INDEX IF NOT EXISTS idx_order_tags_tag ON order_tags(tag);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS order_tags;

DROP INDEX IF EXISTS idx_orders_parent_order_id;

ALTER TABLE orders DROP COLUMN IF EXISTS parent_order_id;

-- +goose StatementEnd
