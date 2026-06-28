-- +goose Up
CREATE TABLE IF NOT EXISTS user_favorite_categories (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_user_favorite_categories_user_id ON user_favorite_categories(user_id);

CREATE TABLE IF NOT EXISTS user_product_views (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_user_product_views_user_viewed ON user_product_views(user_id, viewed_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_product_views_product_id ON user_product_views(product_id);

CREATE TABLE IF NOT EXISTS flash_deals (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    ends_at TIMESTAMPTZ NOT NULL,
    quantity_limit INT,
    sort_order INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flash_deals_status_ends ON flash_deals(status, ends_at);

CREATE TABLE IF NOT EXISTS homepage_sections (
    id BIGSERIAL PRIMARY KEY,
    section_key VARCHAR(64) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    href VARCHAR(512) NOT NULL DEFAULT '/shop',
    image_url VARCHAR(512) NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    filters JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_homepage_sections_status_sort ON homepage_sections(status, sort_order);

CREATE INDEX IF NOT EXISTS idx_order_items_product_created ON order_items(product_id, created_at);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_status_rating ON products(status, rating DESC, reviews_count DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_products_status_rating;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_order_items_product_created;
DROP TABLE IF EXISTS homepage_sections;
DROP TABLE IF EXISTS flash_deals;
DROP TABLE IF EXISTS user_product_views;
DROP TABLE IF EXISTS user_favorite_categories;
