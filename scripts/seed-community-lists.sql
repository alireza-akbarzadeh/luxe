-- Dev/staging community shopping list demo data. Run: psql $DATABASE_URL -f scripts/seed-community-lists.sql
-- Requires seed-dev products (pdp-demo-* watches).

DO $$
DECLARE
    list_id BIGINT;
    prod_luxe BIGINT;
    prod_gold BIGINT;
    prod_urban BIGINT;
BEGIN
    SELECT id INTO prod_luxe FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_gold FROM products WHERE slug = 'pdp-demo-gold-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_urban FROM products WHERE slug = 'pdp-demo-urban-watch' AND deleted_at IS NULL;

    IF prod_luxe IS NULL THEN
        RAISE NOTICE 'seed-community-lists: missing pdp-demo products — run seed-dev first';
        RETURN;
    END IF;

    INSERT INTO community_shopping_lists (
        slug, title, description, theme,
        cover_image_url, author_name, author_handle, is_active, sort_order
    )
    VALUES (
        'first-apartment-essentials',
        'First apartment essentials',
        'Everything you need to make a rental feel like home — from the entryway to the bedside table.',
        'Home & living',
        'https://images.unsplash.com/photo-1616486338812-3dadae4b4ace?w=1600',
        'Maya Chen',
        '@mayachen',
        TRUE,
        0
    )
    ON CONFLICT (slug) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        theme = EXCLUDED.theme,
        cover_image_url = EXCLUDED.cover_image_url,
        author_name = EXCLUDED.author_name,
        author_handle = EXCLUDED.author_handle,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO list_id;

    DELETE FROM community_shopping_list_items WHERE list_id = list_id;

    INSERT INTO community_shopping_list_items (list_id, product_id, note, sort_order)
    VALUES (list_id, prod_luxe, 'The statement piece for your entryway console', 0);

    IF prod_gold IS NOT NULL THEN
        INSERT INTO community_shopping_list_items (list_id, product_id, note, sort_order)
        VALUES (list_id, prod_gold, 'Warm gold tones for evening hosting', 1);
    END IF;

    IF prod_urban IS NOT NULL THEN
        INSERT INTO community_shopping_list_items (list_id, product_id, note, sort_order)
        VALUES (list_id, prod_urban, 'Everyday wear for WFH days', 2);
    END IF;

    INSERT INTO community_shopping_lists (
        slug, title, description, theme,
        cover_image_url, author_name, author_handle, is_active, sort_order
    )
    VALUES (
        'weekend-getaway-capsule',
        'Weekend getaway capsule',
        'Pack light, look polished — a carry-on friendly edit for 48-hour escapes.',
        'Travel',
        'https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=1600',
        'Leo Park',
        '@leopark',
        TRUE,
        1
    )
    ON CONFLICT (slug) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        theme = EXCLUDED.theme,
        cover_image_url = EXCLUDED.cover_image_url,
        author_name = EXCLUDED.author_name,
        author_handle = EXCLUDED.author_handle,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO list_id;

    DELETE FROM community_shopping_list_items WHERE list_id = list_id;

    INSERT INTO community_shopping_list_items (list_id, product_id, note, sort_order)
    VALUES (list_id, prod_luxe, 'Versatile day-to-dinner timepiece', 0);

    IF prod_urban IS NOT NULL THEN
        INSERT INTO community_shopping_list_items (list_id, product_id, note, sort_order)
        VALUES (list_id, prod_urban, 'Casual brunches and city walks', 1);
    END IF;
END $$;
