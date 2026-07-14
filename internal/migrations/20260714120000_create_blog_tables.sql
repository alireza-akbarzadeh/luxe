-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS blog_categories (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT         NOT NULL UNIQUE,
    name        TEXT         NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    image_url   TEXT         NOT NULL DEFAULT '',
    parent_id   BIGINT       REFERENCES blog_categories(id) ON DELETE SET NULL,
    sort_order  INT          NOT NULL DEFAULT 0,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blog_categories_parent_id ON blog_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_blog_categories_is_active ON blog_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_blog_categories_deleted_at ON blog_categories(deleted_at);

CREATE TABLE IF NOT EXISTS blog_authors (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT         NOT NULL UNIQUE,
    name        TEXT         NOT NULL,
    role        TEXT         NOT NULL DEFAULT '',
    bio         TEXT         NOT NULL DEFAULT '',
    avatar_url  TEXT         NOT NULL DEFAULT '',
    user_id     BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blog_authors_user_id ON blog_authors(user_id);
CREATE INDEX IF NOT EXISTS idx_blog_authors_deleted_at ON blog_authors(deleted_at);

CREATE TABLE IF NOT EXISTS blog_tags (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT         NOT NULL UNIQUE,
    name        TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blog_tags_deleted_at ON blog_tags(deleted_at);

CREATE TABLE IF NOT EXISTS blog_posts (
    id                  BIGSERIAL PRIMARY KEY,
    slug                TEXT         NOT NULL UNIQUE,
    title               TEXT         NOT NULL,
    excerpt             TEXT         NOT NULL DEFAULT '',
    hero_image_url      TEXT         NOT NULL DEFAULT '',
    hero_image_alt      TEXT         NOT NULL DEFAULT '',
    content_blocks      JSONB        NOT NULL DEFAULT '[]',
    category_id         BIGINT       REFERENCES blog_categories(id) ON DELETE SET NULL,
    author_id           BIGINT       REFERENCES blog_authors(id) ON DELETE SET NULL,
    section_type        TEXT         NOT NULL DEFAULT 'article',
    status              TEXT         NOT NULL DEFAULT 'draft',
    is_featured         BOOLEAN      NOT NULL DEFAULT FALSE,
    is_editor_pick      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_trending         BOOLEAN      NOT NULL DEFAULT FALSE,
    reading_time_minutes INT         NOT NULL DEFAULT 5,
    view_count          BIGINT       NOT NULL DEFAULT 0,
    helpful_votes       BIGINT       NOT NULL DEFAULT 0,
    meta_title          TEXT         NOT NULL DEFAULT '',
    meta_description    TEXT         NOT NULL DEFAULT '',
    canonical_url       TEXT         NOT NULL DEFAULT '',
    seo_score           INT          NOT NULL DEFAULT 0,
    published_at        TIMESTAMPTZ,
    content_updated_at  TIMESTAMPTZ,
    scheduled_at        TIMESTAMPTZ,
    sort_order          INT          NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blog_posts_category_id ON blog_posts(category_id);
CREATE INDEX IF NOT EXISTS idx_blog_posts_author_id ON blog_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_blog_posts_status ON blog_posts(status);
CREATE INDEX IF NOT EXISTS idx_blog_posts_section_type ON blog_posts(section_type);
CREATE INDEX IF NOT EXISTS idx_blog_posts_published_at ON blog_posts(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_blog_posts_is_featured ON blog_posts(is_featured);
CREATE INDEX IF NOT EXISTS idx_blog_posts_is_trending ON blog_posts(is_trending);
CREATE INDEX IF NOT EXISTS idx_blog_posts_deleted_at ON blog_posts(deleted_at);

CREATE TABLE IF NOT EXISTS blog_post_tags (
    post_id     BIGINT NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    tag_id      BIGINT NOT NULL REFERENCES blog_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS blog_post_products (
    id          BIGSERIAL PRIMARY KEY,
    post_id     BIGINT       NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    product_id  BIGINT       NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    block_type  TEXT         NOT NULL DEFAULT 'recommended',
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blog_post_products_post_id ON blog_post_products(post_id);
CREATE INDEX IF NOT EXISTS idx_blog_post_products_product_id ON blog_post_products(product_id);

CREATE TABLE IF NOT EXISTS blog_comments (
    id                      BIGSERIAL PRIMARY KEY,
    post_id                 BIGINT       NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    user_id                 BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id               BIGINT       REFERENCES blog_comments(id) ON DELETE CASCADE,
    content                 TEXT         NOT NULL,
    like_count              INT          NOT NULL DEFAULT 0,
    is_pinned               BOOLEAN      NOT NULL DEFAULT FALSE,
    is_verified_purchaser   BOOLEAN      NOT NULL DEFAULT FALSE,
    status                  TEXT         NOT NULL DEFAULT 'approved',
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_blog_comments_post_id ON blog_comments(post_id);
CREATE INDEX IF NOT EXISTS idx_blog_comments_user_id ON blog_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_blog_comments_parent_id ON blog_comments(parent_id);

CREATE TABLE IF NOT EXISTS blog_bookmarks (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id     BIGINT NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

-- Seed categories
INSERT INTO blog_categories (slug, name, description, sort_order) VALUES
    ('technology', 'Technology', 'Latest tech news, reviews, and insights', 1),
    ('smartphones', 'Smartphones', 'Phone reviews, comparisons, and buying guides', 2),
    ('laptops', 'Laptops', 'Laptop reviews and productivity guides', 3),
    ('gaming', 'Gaming', 'Gaming gear, reviews, and setup guides', 4),
    ('accessories', 'Accessories', 'Essential accessories and gear', 5),
    ('wearables', 'Wearables', 'Smartwatches, fitness trackers, and more', 6),
    ('photography', 'Photography', 'Cameras, lenses, and creative tools', 7),
    ('audio', 'Audio', 'Headphones, speakers, and sound quality', 8),
    ('home-electronics', 'Home Electronics', 'Smart home and living room tech', 9),
    ('smart-home', 'Smart Home', 'Connected home devices and automation', 10),
    ('office-setup', 'Office Setup', 'Work-from-home and productivity setups', 11),
    ('artificial-intelligence', 'Artificial Intelligence', 'AI tools, trends, and applications', 12),
    ('buying-guides', 'Buying Guides', 'Expert recommendations for every budget', 13),
    ('product-reviews', 'Product Reviews', 'In-depth, honest product reviews', 14),
    ('comparisons', 'Comparisons', 'Side-by-side product comparisons', 15),
    ('how-to-guides', 'How-To Guides', 'Step-by-step tutorials and walkthroughs', 16),
    ('troubleshooting', 'Troubleshooting', 'Fix common issues and problems', 17),
    ('tips-tricks', 'Tips & Tricks', 'Quick tips to get more from your gear', 18),
    ('industry-news', 'Industry News', 'Breaking news and market updates', 19),
    ('lifestyle', 'Lifestyle', 'Tech meets everyday living', 20),
    ('fashion', 'Fashion', 'Style, wearables, and luxury accessories', 21),
    ('luxury', 'Luxury', 'Premium products and exclusive edits', 22),
    ('gift-ideas', 'Gift Ideas', 'Curated gift guides for every occasion', 23),
    ('seasonal-collections', 'Seasonal Collections', 'Seasonal picks and holiday edits', 24)
ON CONFLICT (slug) DO NOTHING;

-- Seed authors
INSERT INTO blog_authors (slug, name, role, bio, avatar_url) VALUES
    ('sarah-chen', 'Sarah Chen', 'Senior Tech Editor', 'Sarah has spent over a decade reviewing consumer electronics. Her work has been featured in major publications and she specializes in smartphones and wearables.', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&h=200&fit=crop'),
    ('marcus-williams', 'Marcus Williams', 'Buying Guide Specialist', 'Marcus helps readers make confident purchase decisions with data-driven comparisons and real-world testing.', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=200&h=200&fit=crop'),
    ('elena-rodriguez', 'Elena Rodriguez', 'Lifestyle & Luxury Editor', 'Elena curates premium lifestyle content and seasonal gift guides for the discerning shopper.', 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&h=200&fit=crop')
ON CONFLICT (slug) DO NOTHING;

-- Seed tags
INSERT INTO blog_tags (slug, name) VALUES
    ('iphone', 'iPhone'), ('android', 'Android'), ('macbook', 'MacBook'),
    ('wireless', 'Wireless'), ('premium', 'Premium'), ('budget', 'Budget'),
    ('2026', '2026'), ('gift-guide', 'Gift Guide'), ('black-friday', 'Black Friday')
ON CONFLICT (slug) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blog_bookmarks;
DROP TABLE IF EXISTS blog_comments;
DROP TABLE IF EXISTS blog_post_products;
DROP TABLE IF EXISTS blog_post_tags;
DROP TABLE IF EXISTS blog_posts;
DROP TABLE IF EXISTS blog_tags;
DROP TABLE IF EXISTS blog_authors;
DROP TABLE IF EXISTS blog_categories;
-- +goose StatementEnd
