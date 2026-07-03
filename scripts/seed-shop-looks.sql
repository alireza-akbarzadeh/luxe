-- Dev/staging shop-the-look demo scenes. Run: psql $DATABASE_URL -f scripts/seed-shop-looks.sql
-- Requires seed-dev products (pdp-demo-* watches).

DO $$
DECLARE
    look_id BIGINT;
    prod_luxe BIGINT;
    prod_gold BIGINT;
    prod_urban BIGINT;
BEGIN
    SELECT id INTO prod_luxe FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_gold FROM products WHERE slug = 'pdp-demo-gold-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_urban FROM products WHERE slug = 'pdp-demo-urban-watch' AND deleted_at IS NULL;

    IF prod_luxe IS NULL THEN
        RAISE NOTICE 'seed-shop-looks: missing pdp-demo products — run seed-dev first';
        RETURN;
    END IF;

    INSERT INTO shop_looks (slug, title, description, image_url, is_active, sort_order)
    VALUES (
        'executive-desk-edit',
        'The executive desk edit',
        'A considered workspace pairing — heritage chronograph, leather accents, and warm lighting for focused days.',
        'https://images.unsplash.com/photo-1616486338812-3dadae4b4ace?w=1600',
        TRUE,
        0
    )
    ON CONFLICT (slug) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        image_url = EXCLUDED.image_url,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO look_id;

    DELETE FROM shop_look_tags WHERE shop_look_id = look_id;

    INSERT INTO shop_look_tags (shop_look_id, product_id, x_percent, y_percent, label, sort_order)
    VALUES
        (look_id, prod_luxe, 62.0, 58.0, 'Heritage chronograph', 0);

    IF prod_gold IS NOT NULL THEN
        INSERT INTO shop_look_tags (shop_look_id, product_id, x_percent, y_percent, label, sort_order)
        VALUES (look_id, prod_gold, 28.0, 72.0, 'Gold market pick', 1);
    END IF;

    IF prod_urban IS NOT NULL THEN
        INSERT INTO shop_look_tags (shop_look_id, product_id, x_percent, y_percent, label, sort_order)
        VALUES (look_id, prod_urban, 78.0, 38.0, 'Urban essentials', 2);
    END IF;

    INSERT INTO shop_looks (slug, title, description, image_url, is_active, sort_order)
    VALUES (
        'evening-lobby-look',
        'Evening lobby look',
        'Dress the part for an evening out — tap each piece to shop the full edit.',
        'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1600',
        TRUE,
        1
    )
    ON CONFLICT (slug) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        image_url = EXCLUDED.image_url,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO look_id;

    DELETE FROM shop_look_tags WHERE shop_look_id = look_id;

    INSERT INTO shop_look_tags (shop_look_id, product_id, x_percent, y_percent, label, sort_order)
    VALUES (look_id, prod_luxe, 45.0, 55.0, 'Statement timepiece', 0);

    RAISE NOTICE 'seed-shop-looks: demo scenes ready';
END $$;
