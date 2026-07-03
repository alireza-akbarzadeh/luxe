-- +goose Up
-- +goose StatementBegin

INSERT INTO stores (name, slug, description, logo_url, banner_url, is_verified, rating, review_count, follower_count, location, shipping_info, return_policy, status)
VALUES
  ('Luxe Atelier', 'luxe-atelier', 'Curated luxury fashion and accessories with white-glove service.', 'https://images.unsplash.com/photo-1560179707-f14e90ef3623?w=200', 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=1200', true, 4.8, 128, 2400, 'Milan, Italy', 'Complimentary express shipping on orders over $150. Delivery in 2–4 business days.', '30-day returns on unworn items with tags attached.', 'active'),
  ('Gold Market', 'gold-market', 'Premium timepieces and jewelry from trusted sellers.', 'https://images.unsplash.com/photo-1612817159949-1957456bae79?w=200', 'https://images.unsplash.com/photo-1515562141207-29a036fb126a?w=1200', true, 4.6, 89, 1800, 'Zurich, Switzerland', 'Insured shipping worldwide. Signature required on delivery.', '14-day exchange policy on unworn watches.', 'active'),
  ('Urban Essentials', 'urban-essentials', 'Modern staples with fast domestic fulfillment.', 'https://images.unsplash.com/photo-1472851294608-062f824d29cc?w=200', 'https://images.unsplash.com/photo-1555529669-2269763671c0?w=1200', false, 4.3, 54, 920, 'New York, USA', 'Standard shipping 3–5 days. Free over $99.', 'Returns accepted within 21 days.', 'active')
ON CONFLICT (slug) DO NOTHING;

DO $$
DECLARE
    cat_id INT;
    store_luxe INT;
    store_gold INT;
    store_urban INT;
    prod_main INT;
    prod_oos INT;
    user_id INT;
BEGIN
    SELECT id INTO cat_id FROM categories WHERE deleted_at IS NULL ORDER BY id LIMIT 1;
    SELECT id INTO store_luxe FROM stores WHERE slug = 'luxe-atelier';
    SELECT id INTO store_gold FROM stores WHERE slug = 'gold-market';
    SELECT id INTO store_urban FROM stores WHERE slug = 'urban-essentials';
    SELECT id INTO user_id FROM users ORDER BY id LIMIT 1;

    IF cat_id IS NULL OR store_luxe IS NULL THEN
        RETURN;
    END IF;

    INSERT INTO products (
        name, slug, description, price, compare_at_price, stock, sku, barcode, category_id, store_id,
        status, rating, reviews_count, is_new, images, colors, sizes, tags, visibility,
        track_inventory, allow_backorder, weight
    ) VALUES (
        'Heritage Chronograph Watch',
        'pdp-demo-luxe-watch',
        'A refined chronograph with sapphire crystal, Swiss movement, and a sunburst dial. Water resistant to 100m. Includes presentation box and 2-year warranty.',
        1299.00, 1499.00, 12, 'PDP-WATCH-LUXE-001', '8801234567890', cat_id, store_luxe,
        'active', 4.7, 18, true,
        ARRAY['https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800','https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=800'],
        '["Silver","Gold"]'::jsonb, '["40mm","42mm"]'::jsonb,
        ARRAY['watch','chronograph','luxury','Swiss'],
        'public', true, false, 0.42
    ) ON CONFLICT (slug) DO NOTHING;

    INSERT INTO products (
        name, slug, description, price, compare_at_price, stock, sku, barcode, category_id, store_id,
        status, rating, reviews_count, images, tags, visibility, track_inventory, weight
    ) VALUES (
        'Heritage Chronograph Watch',
        'pdp-demo-gold-watch',
        'Same Heritage Chronograph model from Gold Market — authenticated and insured shipping.',
        1249.00, 1399.00, 5, 'PDP-WATCH-GOLD-001', '8801234567890', cat_id, store_gold,
        'active', 4.5, 11,
        ARRAY['https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800'],
        ARRAY['watch','chronograph','marketplace'],
        'public', true, 0.42
    ) ON CONFLICT (slug) DO NOTHING;

    INSERT INTO products (
        name, slug, description, price, stock, sku, barcode, category_id, store_id,
        status, rating, reviews_count, images, tags, visibility, track_inventory, weight
    ) VALUES (
        'Heritage Chronograph Watch',
        'pdp-demo-urban-watch',
        'Heritage Chronograph from Urban Essentials — best value with fast US shipping.',
        1199.00, 8, 'PDP-WATCH-URBAN-001', '8801234567890', cat_id, store_urban,
        'active', 4.2, 7,
        ARRAY['https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=800'],
        ARRAY['watch','value'],
        'public', true, 0.42
    ) ON CONFLICT (slug) DO NOTHING;

    INSERT INTO products (
        name, slug, description, price, compare_at_price, stock, sku, barcode, category_id, store_id,
        status, rating, reviews_count, images, tags, visibility, track_inventory, weight
    ) VALUES (
        'Heritage Chronograph — Limited Edition',
        'pdp-demo-watch-sold-out',
        'Limited dial variant — currently sold out. Subscribe for back-in-stock alerts.',
        1399.00, 1599.00, 0, 'PDP-WATCH-OOS-001', '8801234567891', cat_id, store_luxe,
        'active', 4.9, 24,
        ARRAY['https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800'],
        ARRAY['watch','limited','sold-out'],
        'public', true, 0.44
    ) ON CONFLICT (slug) DO NOTHING;

    SELECT id INTO prod_main FROM products WHERE slug = 'pdp-demo-luxe-watch';
    SELECT id INTO prod_oos FROM products WHERE slug = 'pdp-demo-watch-sold-out';

    IF prod_main IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM product_attributes WHERE product_id = prod_main LIMIT 1) THEN
            INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
            VALUES
                (prod_main, 'color', ARRAY['Silver', 'Gold'], NOW(), NOW()),
                (prod_main, 'size', ARRAY['40mm', '42mm'], NOW(), NOW()),
                (prod_main, 'Material', ARRAY['Stainless steel', 'Sapphire crystal'], NOW(), NOW()),
                (prod_main, 'Movement', ARRAY['Swiss automatic'], NOW(), NOW()),
                (prod_main, 'Water resistance', ARRAY['100m'], NOW(), NOW()),
                (prod_main, 'video', ARRAY['https://www.youtube.com/watch?v=1La4QzGe55Q'], NOW(), NOW());
        END IF;

        INSERT INTO product_price_history (product_id, price, compare_at_price, recorded_at)
        SELECT prod_main, vals.price_val, vals.compare_val, vals.ts
        FROM (
            VALUES
                (1499.00::decimal, 1699.00::decimal, NOW() - INTERVAL '84 days'),
                (1449.00::decimal, 1699.00::decimal, NOW() - INTERVAL '77 days'),
                (1399.00::decimal, 1599.00::decimal, NOW() - INTERVAL '70 days'),
                (1379.00::decimal, 1599.00::decimal, NOW() - INTERVAL '63 days'),
                (1349.00::decimal, 1549.00::decimal, NOW() - INTERVAL '56 days'),
                (1329.00::decimal, 1549.00::decimal, NOW() - INTERVAL '49 days'),
                (1319.00::decimal, 1499.00::decimal, NOW() - INTERVAL '42 days'),
                (1309.00::decimal, 1499.00::decimal, NOW() - INTERVAL '35 days'),
                (1299.00::decimal, 1499.00::decimal, NOW() - INTERVAL '28 days'),
                (1299.00::decimal, 1499.00::decimal, NOW() - INTERVAL '21 days'),
                (1279.00::decimal, 1449.00::decimal, NOW() - INTERVAL '14 days'),
                (1299.00::decimal, 1499.00::decimal, NOW() - INTERVAL '7 days'),
                (1299.00::decimal, 1499.00::decimal, NOW())
        ) AS vals(price_val, compare_val, ts)
        WHERE NOT EXISTS (
            SELECT 1 FROM product_price_history WHERE product_id = prod_main LIMIT 1
        );

        IF user_id IS NOT NULL AND NOT EXISTS (
            SELECT 1 FROM product_questions WHERE product_id = prod_main LIMIT 1
        ) THEN
            INSERT INTO product_questions (product_id, user_id, body, created_at, updated_at)
            VALUES
                (prod_main, user_id, 'Is this watch suitable for swimming?', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
                (prod_main, user_id, 'What is the return policy if the size does not fit?', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days');
        END IF;
    END IF;

    IF prod_oos IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM product_price_history WHERE product_id = prod_oos LIMIT 1
    ) THEN
        INSERT INTO product_price_history (product_id, price, compare_at_price, recorded_at)
        VALUES
            (prod_oos, 1499.00, 1699.00, NOW() - INTERVAL '30 days'),
            (prod_oos, 1449.00, 1599.00, NOW() - INTERVAL '15 days'),
            (prod_oos, 1399.00, 1599.00, NOW());
    END IF;
END $$;

INSERT INTO product_answers (question_id, user_id, body, is_store_reply, is_ai_reply, created_at, updated_at)
SELECT q.id, COALESCE(s.user_id, q.user_id), 'Thanks for asking! This model is water resistant to 100m — suitable for swimming and daily wear.', false, true, q.created_at + INTERVAL '1 hour', q.created_at + INTERVAL '1 hour'
FROM product_questions q
JOIN products p ON p.id = q.product_id
JOIN stores s ON s.id = p.store_id
WHERE p.slug = 'pdp-demo-luxe-watch'
  AND q.body ILIKE '%swim%'
  AND NOT EXISTS (SELECT 1 FROM product_answers a WHERE a.question_id = q.id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM product_answers WHERE question_id IN (
    SELECT id FROM product_questions WHERE product_id IN (
        SELECT id FROM products WHERE slug LIKE 'pdp-demo-%'
    )
);
DELETE FROM product_questions WHERE product_id IN (
    SELECT id FROM products WHERE slug LIKE 'pdp-demo-%'
);
DELETE FROM product_price_history WHERE product_id IN (
    SELECT id FROM products WHERE slug LIKE 'pdp-demo-%'
);
DELETE FROM product_attributes WHERE product_id IN (
    SELECT id FROM products WHERE slug LIKE 'pdp-demo-%'
);
DELETE FROM stock_notifications WHERE product_id IN (
    SELECT id FROM products WHERE slug LIKE 'pdp-demo-%'
);
DELETE FROM products WHERE slug LIKE 'pdp-demo-%';
DELETE FROM stores WHERE slug IN ('luxe-atelier', 'gold-market', 'urban-essentials');

-- +goose StatementEnd
