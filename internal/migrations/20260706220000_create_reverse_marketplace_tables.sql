-- +goose Up
CREATE TABLE IF NOT EXISTS reverse_marketplace_requests (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT           NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    category    TEXT           NOT NULL DEFAULT '',
    budget_min  DECIMAL(10, 2) NULL CHECK (budget_min >= 0),
    budget_max  DECIMAL(10, 2) NULL CHECK (budget_max >= 0),
    status      TEXT           NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'fulfilled')),
    expires_at  TIMESTAMPTZ    NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ    NULL
);

CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_requests_user_id ON reverse_marketplace_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_requests_status ON reverse_marketplace_requests(status);
CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_requests_deleted_at ON reverse_marketplace_requests(deleted_at);

CREATE TABLE IF NOT EXISTS reverse_marketplace_offers (
    id              BIGSERIAL PRIMARY KEY,
    request_id      BIGINT         NOT NULL REFERENCES reverse_marketplace_requests(id) ON DELETE CASCADE,
    store_id        BIGINT         NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    vendor_user_id  BIGINT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message         TEXT           NOT NULL DEFAULT '',
    offered_price   DECIMAL(10, 2) NOT NULL CHECK (offered_price >= 0),
    status          TEXT           NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'withdrawn')),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ    NULL,
    UNIQUE (request_id, store_id)
);

CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_offers_request_id ON reverse_marketplace_offers(request_id);
CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_offers_store_id ON reverse_marketplace_offers(store_id);
CREATE INDEX IF NOT EXISTS idx_reverse_marketplace_offers_deleted_at ON reverse_marketplace_offers(deleted_at);

-- +goose Down
DROP TABLE IF EXISTS reverse_marketplace_offers;
DROP TABLE IF EXISTS reverse_marketplace_requests;
