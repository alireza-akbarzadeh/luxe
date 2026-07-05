-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS creators (
    id              BIGSERIAL PRIMARY KEY,
    slug            TEXT         NOT NULL UNIQUE,
    display_name    TEXT         NOT NULL,
    handle          TEXT         NOT NULL DEFAULT '',
    bio             TEXT         NOT NULL DEFAULT '',
    specialty       TEXT         NOT NULL DEFAULT '',
    avatar_url      TEXT         NOT NULL DEFAULT '',
    cover_image_url TEXT         NOT NULL DEFAULT '',
    instagram_url   TEXT         NOT NULL DEFAULT '',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_creators_is_active ON creators(is_active);
CREATE INDEX IF NOT EXISTS idx_creators_sort_order ON creators(sort_order);
CREATE INDEX IF NOT EXISTS idx_creators_deleted_at ON creators(deleted_at);

CREATE TABLE IF NOT EXISTS creator_picks (
    id          BIGSERIAL PRIMARY KEY,
    creator_id  BIGINT       NOT NULL REFERENCES creators(id) ON DELETE CASCADE,
    product_id  BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    headline    TEXT         NOT NULL DEFAULT '',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_creator_picks_creator_id ON creator_picks(creator_id);
CREATE INDEX IF NOT EXISTS idx_creator_picks_product_id ON creator_picks(product_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS creator_picks;
DROP TABLE IF EXISTS creators;
-- +goose StatementEnd
