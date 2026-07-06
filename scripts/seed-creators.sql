-- Dev/staging creator storefront demo profiles. Run: psql $DATABASE_URL -f scripts/seed-creators.sql
-- Requires seed-dev products (pdp-demo-* watches).

DO $$
DECLARE
    v_creator_id BIGINT;
    prod_luxe BIGINT;
    prod_gold BIGINT;
    prod_urban BIGINT;
BEGIN
    SELECT id INTO prod_luxe FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_gold FROM products WHERE slug = 'pdp-demo-gold-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_urban FROM products WHERE slug = 'pdp-demo-urban-watch' AND deleted_at IS NULL;

    IF prod_luxe IS NULL THEN
        RAISE NOTICE 'seed-creators: missing pdp-demo products — run seed-dev first';
        RETURN;
    END IF;

    INSERT INTO creators (
        slug, display_name, handle, bio, specialty,
        avatar_url, cover_image_url, instagram_url, is_active, sort_order
    )
    VALUES (
        'elena-marchetti',
        'Elena Marchetti',
        '@elenamstyle',
        'Luxury stylist and watch collector sharing timeless edits for modern wardrobes.',
        'Heritage timepieces & evening wear',
        'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=400&h=400&fit=crop',
        'https://images.unsplash.com/photo-1519741497674-611481863552?w=1600',
        'https://instagram.com/elenamstyle',
        TRUE,
        0
    )
    ON CONFLICT (slug) DO UPDATE SET
        display_name = EXCLUDED.display_name,
        handle = EXCLUDED.handle,
        bio = EXCLUDED.bio,
        specialty = EXCLUDED.specialty,
        avatar_url = EXCLUDED.avatar_url,
        cover_image_url = EXCLUDED.cover_image_url,
        instagram_url = EXCLUDED.instagram_url,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO v_creator_id;

    DELETE FROM creator_picks cp WHERE cp.creator_id = v_creator_id;

    INSERT INTO creator_picks (creator_id, product_id, headline, sort_order)
    VALUES (v_creator_id, prod_luxe, 'My everyday heritage pick', 0);

    IF prod_gold IS NOT NULL THEN
        INSERT INTO creator_picks (creator_id, product_id, headline, sort_order)
        VALUES (v_creator_id, prod_gold, 'Statement piece for events', 1);
    END IF;

    IF prod_urban IS NOT NULL THEN
        INSERT INTO creator_picks (creator_id, product_id, headline, sort_order)
        VALUES (v_creator_id, prod_urban, 'Weekend casual rotation', 2);
    END IF;

    INSERT INTO creators (
        slug, display_name, handle, bio, specialty,
        avatar_url, cover_image_url, instagram_url, is_active, sort_order
    )
    VALUES (
        'james-cole',
        'James Cole',
        '@jamescole',
        'Minimalist menswear curator — fewer pieces, better stories.',
        'Minimal luxury & desk-to-dinner',
        'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&h=400&fit=crop',
        'https://images.unsplash.com/photo-1616486338812-3dadae4b4ace?w=1600',
        'https://instagram.com/jamescole',
        TRUE,
        1
    )
    ON CONFLICT (slug) DO UPDATE SET
        display_name = EXCLUDED.display_name,
        handle = EXCLUDED.handle,
        bio = EXCLUDED.bio,
        specialty = EXCLUDED.specialty,
        avatar_url = EXCLUDED.avatar_url,
        cover_image_url = EXCLUDED.cover_image_url,
        instagram_url = EXCLUDED.instagram_url,
        is_active = EXCLUDED.is_active,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO v_creator_id;

    DELETE FROM creator_picks cp WHERE cp.creator_id = v_creator_id;

    IF prod_urban IS NOT NULL THEN
        INSERT INTO creator_picks (creator_id, product_id, headline, sort_order)
        VALUES (v_creator_id, prod_urban, 'Desk-to-dinner essential', 0);
    END IF;

    IF prod_luxe IS NOT NULL THEN
        INSERT INTO creator_picks (creator_id, product_id, headline, sort_order)
        VALUES (v_creator_id, prod_luxe, 'The one I recommend most', 1);
    END IF;

    RAISE NOTICE 'seed-creators: demo creator storefronts ready';
END $$;
