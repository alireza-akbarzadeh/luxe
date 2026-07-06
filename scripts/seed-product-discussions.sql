-- Dev/staging product discussion demo data. Run: psql $DATABASE_URL -f scripts/seed-product-discussions.sql
-- Requires seed-dev products (pdp-demo-* watches).

DO $$
DECLARE
    prod_luxe BIGINT;
    user_id BIGINT;
    disc_id BIGINT;
BEGIN
    SELECT id INTO prod_luxe FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL;
    SELECT id INTO user_id FROM users ORDER BY id LIMIT 1;

    IF prod_luxe IS NULL OR user_id IS NULL THEN
        RAISE NOTICE 'seed-product-discussions: missing pdp-demo product or user — run seed-dev first';
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM product_discussions WHERE product_id = prod_luxe LIMIT 1) THEN
        RAISE NOTICE 'seed-product-discussions: discussions already seeded for luxe watch';
        RETURN;
    END IF;

    INSERT INTO product_discussions (product_id, user_id, title, body, created_at, updated_at)
    VALUES (
        prod_luxe,
        user_id,
        'Daily wear vs dress occasions',
        'I am torn between using this as an everyday piece or saving it for events. How does the bracelet hold up to daily desk work?',
        NOW() - INTERVAL '6 days',
        NOW() - INTERVAL '6 days'
    )
    RETURNING id INTO disc_id;

    INSERT INTO product_discussion_replies (discussion_id, user_id, body, created_at, updated_at)
    VALUES
        (
            disc_id,
            user_id,
            'I have worn mine daily for three months — minor desk scuffs on the clasp but the crystal is still flawless.',
            NOW() - INTERVAL '5 days',
            NOW() - INTERVAL '5 days'
        ),
        (
            disc_id,
            user_id,
            'Same here. The sunburst dial dresses it up enough for dinner without feeling too flashy at the office.',
            NOW() - INTERVAL '4 days',
            NOW() - INTERVAL '4 days'
        );

    INSERT INTO product_discussions (product_id, user_id, title, body, created_at, updated_at)
    VALUES (
        prod_luxe,
        user_id,
        'Strap sizing for smaller wrists',
        'Has anyone swapped the stock bracelet for a smaller wrist? Curious what lug width and aftermarket options people like.',
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '2 days'
    );
END $$;
