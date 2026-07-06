-- Dev/staging public collection demo data. Run: psql $DATABASE_URL -f scripts/seed-public-collections.sql
-- Requires seed-dev products (pdp-demo-* watches).

DO $$
DECLARE
    collection_id BIGINT;
    prod_luxe BIGINT;
    prod_gold BIGINT;
    prod_urban BIGINT;
BEGIN
    SELECT id INTO prod_luxe FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_gold FROM products WHERE slug = 'pdp-demo-gold-watch' AND deleted_at IS NULL;
    SELECT id INTO prod_urban FROM products WHERE slug = 'pdp-demo-urban-watch' AND deleted_at IS NULL;

    IF prod_luxe IS NULL THEN
        RAISE NOTICE 'seed-public-collections: missing pdp-demo products — run seed-dev first';
        RETURN;
    END IF;

    INSERT INTO public_collections (
        slug, title, description, theme,
        cover_image_url, author_name, author_handle, is_active, sort_order
    )
    VALUES (
        'desk-to-dinner-edit',
        'Desk to dinner edit',
        'Pieces that transition from focused workdays to evening plans without a wardrobe change.',
        'Workwear',
        'https://images.unsplash.com/photo-1497366216548-37526070297c?w=1600',
        'Sofia Alvarez',
        '@sofiaalvarez',
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
    RETURNING id INTO collection_id;

    DELETE FROM public_collection_items WHERE collection_id = collection_id;

    INSERT INTO public_collection_items (collection_id, product_id, note, sort_order)
    VALUES (collection_id, prod_luxe, 'Understated luxury for client calls and dinners', 0);

    IF prod_gold IS NOT NULL THEN
        INSERT INTO public_collection_items (collection_id, product_id, note, sort_order)
        VALUES (collection_id, prod_gold, 'Adds warmth to neutral tailoring', 1);
    END IF;

    INSERT INTO public_collections (
        slug, title, description, theme,
        cover_image_url, author_name, author_handle, is_active, sort_order
    )
    VALUES (
        'gift-guide-heritage',
        'Heritage gift guide',
        'Timeless pieces that make memorable gifts — curated for milestones and celebrations.',
        'Gifting',
        'https://images.unsplash.com/photo-1513885535751-8b9238bd345a?w=1600',
        'Daniel Okonkwo',
        '@danokonkwo',
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
    RETURNING id INTO collection_id;

    DELETE FROM public_collection_items WHERE collection_id = collection_id;

    INSERT INTO public_collection_items (collection_id, product_id, note, sort_order)
    VALUES (collection_id, prod_luxe, 'A classic that never misses', 0);

    IF prod_urban IS NOT NULL THEN
        INSERT INTO public_collection_items (collection_id, product_id, note, sort_order)
        VALUES (collection_id, prod_urban, 'Great for younger recipients starting their collection', 1);
    END IF;
END $$;
