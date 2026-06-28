-- Homepage personalization demo data (favorite categories, wishlist likes).
-- Run after seed-catalog.sql and seed-orders-returns.sql. Idempotent.

DO $$
DECLARE
    buyer_user_id BIGINT;
    cat_women_id BIGINT;
    cat_men_id BIGINT;
    cat_accessories_id BIGINT;
    cat_watches_id BIGINT;
    prod_id BIGINT;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'user_favorite_categories'
    ) THEN
        RAISE NOTICE 'seed-home-personalization: user_favorite_categories missing — run make migrate-up first';
        RETURN;
    END IF;

    SELECT id INTO buyer_user_id FROM users ORDER BY id ASC OFFSET 1 LIMIT 1;
    IF buyer_user_id IS NULL THEN
        SELECT id INTO buyer_user_id FROM users ORDER BY id ASC LIMIT 1;
    END IF;

    IF buyer_user_id IS NULL THEN
        RAISE NOTICE 'seed-home-personalization: no users found';
        RETURN;
    END IF;

    SELECT id INTO cat_women_id FROM categories WHERE slug = 'women' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO cat_men_id FROM categories WHERE slug = 'men' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO cat_accessories_id FROM categories WHERE slug = 'accessories' AND deleted_at IS NULL LIMIT 1;
    SELECT id INTO cat_watches_id FROM categories WHERE slug = 'watches' AND deleted_at IS NULL LIMIT 1;

    -- Explicit favorite categories for the demo buyer
    IF cat_women_id IS NOT NULL THEN
        INSERT INTO user_favorite_categories (user_id, category_id, created_at)
        VALUES (buyer_user_id, cat_women_id, NOW())
        ON CONFLICT (user_id, category_id) DO NOTHING;
    END IF;

    IF cat_accessories_id IS NOT NULL THEN
        INSERT INTO user_favorite_categories (user_id, category_id, created_at)
        VALUES (buyer_user_id, cat_accessories_id, NOW())
        ON CONFLICT (user_id, category_id) DO NOTHING;
    END IF;

    IF cat_men_id IS NOT NULL THEN
        INSERT INTO user_favorite_categories (user_id, category_id, created_at)
        VALUES (buyer_user_id, cat_men_id, NOW())
        ON CONFLICT (user_id, category_id) DO NOTHING;
    END IF;

    -- Wishlist likes across categories (supports derived personalization when favorites are cleared)
    FOR prod_id IN
        SELECT p.id
        FROM products p
        WHERE p.deleted_at IS NULL
          AND p.status = 'active'
          AND p.category_id IS NOT NULL
        ORDER BY p.rating DESC NULLS LAST, p.id ASC
        LIMIT 6
    LOOP
        INSERT INTO product_likes (user_id, product_id, created_at, updated_at)
        VALUES (buyer_user_id, prod_id, NOW(), NOW())
        ON CONFLICT (user_id, product_id) DO NOTHING;
    END LOOP;

    -- Extra like in watches for category diversity
    IF cat_watches_id IS NOT NULL THEN
        SELECT p.id INTO prod_id
        FROM products p
        WHERE p.category_id = cat_watches_id
          AND p.deleted_at IS NULL
          AND p.status = 'active'
        ORDER BY p.id ASC
        LIMIT 1;

        IF prod_id IS NOT NULL THEN
            INSERT INTO product_likes (user_id, product_id, created_at, updated_at)
            VALUES (buyer_user_id, prod_id, NOW(), NOW())
            ON CONFLICT (user_id, product_id) DO NOTHING;
        END IF;
    END IF;
END $$;
