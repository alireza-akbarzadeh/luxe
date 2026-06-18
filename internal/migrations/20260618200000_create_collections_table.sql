-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS collections (
    id                  BIGSERIAL PRIMARY KEY,
    slug                TEXT         NOT NULL UNIQUE,
    eyebrow             TEXT         NOT NULL DEFAULT '',
    title               TEXT         NOT NULL,
    description         TEXT         NOT NULL DEFAULT '',
    href                TEXT         NOT NULL DEFAULT '/shop',
    image_url           TEXT         NOT NULL DEFAULT '',
    cta_label           TEXT         NOT NULL DEFAULT 'Shop collection',
    sort_order          INT          NOT NULL DEFAULT 0,
    status              TEXT         NOT NULL DEFAULT 'draft',
    preview_sort        TEXT         NOT NULL DEFAULT '',
    preview_is_new      BOOLEAN,
    preview_category_id BIGINT       REFERENCES categories(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_collections_status ON collections(status);
CREATE INDEX IF NOT EXISTS idx_collections_sort_order ON collections(sort_order);
CREATE INDEX IF NOT EXISTS idx_collections_deleted_at ON collections(deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collections;
-- +goose StatementEnd
