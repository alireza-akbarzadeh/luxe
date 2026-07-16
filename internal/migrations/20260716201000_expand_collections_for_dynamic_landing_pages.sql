-- +goose Up
-- +goose StatementBegin

ALTER TABLE collections
    ADD COLUMN IF NOT EXISTS subtitle TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'dynamic',
    ADD COLUMN IF NOT EXISTS sort_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS rules_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS seo_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS seo_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS meta_keywords TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS og_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS og_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS og_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS twitter_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS twitter_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS twitter_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS canonical_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS robots_directives TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS is_indexable BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS hero_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS hero_description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS desktop_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tablet_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mobile_image_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS overlay_opacity DOUBLE PRECISION NOT NULL DEFAULT 0.25,
    ADD COLUMN IF NOT EXISTS theme_variant TEXT NOT NULL DEFAULT '';

UPDATE collections
SET mode = CASE
    WHEN collection_type = 'manual' THEN 'manual'
    ELSE 'dynamic'
END
WHERE mode = 'dynamic';

CREATE INDEX IF NOT EXISTS idx_collections_mode ON collections(mode);

ALTER TABLE collection_products
    ADD COLUMN IF NOT EXISTS position INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS boost_score INT NOT NULL DEFAULT 0;

UPDATE collection_products
SET position = sort_order
WHERE position = 0;

CREATE TABLE IF NOT EXISTS collection_slug_redirects (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    old_slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collection_slug_redirects_collection_id ON collection_slug_redirects(collection_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS collection_slug_redirects;

ALTER TABLE collection_products
    DROP COLUMN IF EXISTS boost_score,
    DROP COLUMN IF EXISTS is_hidden,
    DROP COLUMN IF EXISTS is_pinned,
    DROP COLUMN IF EXISTS position;

DROP INDEX IF EXISTS idx_collections_mode;

ALTER TABLE collections
    DROP COLUMN IF EXISTS theme_variant,
    DROP COLUMN IF EXISTS overlay_opacity,
    DROP COLUMN IF EXISTS mobile_image_url,
    DROP COLUMN IF EXISTS tablet_image_url,
    DROP COLUMN IF EXISTS desktop_image_url,
    DROP COLUMN IF EXISTS hero_description,
    DROP COLUMN IF EXISTS hero_title,
    DROP COLUMN IF EXISTS is_indexable,
    DROP COLUMN IF EXISTS robots_directives,
    DROP COLUMN IF EXISTS canonical_url,
    DROP COLUMN IF EXISTS twitter_image_url,
    DROP COLUMN IF EXISTS twitter_description,
    DROP COLUMN IF EXISTS twitter_title,
    DROP COLUMN IF EXISTS og_image_url,
    DROP COLUMN IF EXISTS og_description,
    DROP COLUMN IF EXISTS og_title,
    DROP COLUMN IF EXISTS meta_keywords,
    DROP COLUMN IF EXISTS seo_description,
    DROP COLUMN IF EXISTS seo_title,
    DROP COLUMN IF EXISTS rules_json,
    DROP COLUMN IF EXISTS sort_key,
    DROP COLUMN IF EXISTS mode,
    DROP COLUMN IF EXISTS subtitle;

-- +goose StatementEnd
