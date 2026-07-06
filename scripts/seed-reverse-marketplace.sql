-- Dev/staging reverse marketplace demo data. Run: psql $DATABASE_URL -f scripts/seed-reverse-marketplace.sql

DO $$
DECLARE
    buyer_id BIGINT;
    store_id BIGINT;
    req_id BIGINT;
BEGIN
    SELECT id INTO buyer_id FROM users ORDER BY id LIMIT 1;
    SELECT id INTO store_id FROM stores WHERE deleted_at IS NULL ORDER BY id LIMIT 1;

    IF buyer_id IS NULL OR store_id IS NULL THEN
        RAISE NOTICE 'seed-reverse-marketplace: missing user or store — run seed-dev first';
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM reverse_marketplace_requests LIMIT 1) THEN
        RAISE NOTICE 'seed-reverse-marketplace: requests already seeded';
        RETURN;
    END IF;

    INSERT INTO reverse_marketplace_requests (user_id, title, description, category, budget_min, budget_max, status, created_at, updated_at)
    VALUES (
        buyer_id,
        'Vintage dress watch under $2,500',
        'Looking for a slim automatic dress watch with a silver dial. Prefer 38–40mm case and exhibition caseback.',
        'Watches',
        1500,
        2500,
        'open',
        NOW() - INTERVAL '3 days',
        NOW() - INTERVAL '3 days'
    )
    RETURNING id INTO req_id;

    INSERT INTO reverse_marketplace_offers (request_id, store_id, vendor_user_id, message, offered_price, status, created_at, updated_at)
    VALUES (
        req_id,
        store_id,
        buyer_id,
        'We have a curated dress watch that matches your size and budget. Includes 2-year warranty.',
        2195,
        'pending',
        NOW() - INTERVAL '2 days',
        NOW() - INTERVAL '2 days'
    );

    INSERT INTO reverse_marketplace_requests (user_id, title, description, category, budget_min, budget_max, status, created_at, updated_at)
    VALUES (
        buyer_id,
        'Minimalist leather weekender bag',
        'Earth-tone leather, fits a 15" laptop, prefer neutral hardware.',
        'Bags',
        200,
        450,
        'open',
        NOW() - INTERVAL '1 day',
        NOW() - INTERVAL '1 day'
    );
END $$;
