-- +goose Up
-- +goose StatementBegin
-- Repair: order_tags / parent_order_id may be missing when goose version was ahead of schema.

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
-- No-op down: table may predate this repair migration on healthy databases.
-- +goose StatementEnd
