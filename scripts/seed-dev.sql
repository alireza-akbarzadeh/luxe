-- Dev/staging demo data. Run: make seed-dev
-- Idempotent where possible. Not for production.

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
    SELECT id INTO cat_id FROM categories WHERE slug = 'watches' AND deleted_at IS NULL;
    IF cat_id IS NULL THEN
        SELECT id INTO cat_id FROM categories WHERE deleted_at IS NULL ORDER BY id LIMIT 1;
    END IF;
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

-- Storefront navigation (nav_menus). One row per label; safe to re-run.
INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Women',
  'mega',
  NULL,
  NULL,
  '{"label":"Shop all women''s","href":"/shop"}'::jsonb,
  '[
    {"title":"Clothing","links":[
      {"title":"Dresses","href":"/shop?sortBy=newest"},
      {"title":"Tops & Blouses","href":"/shop"},
      {"title":"Knitwear","href":"/shop"},
      {"title":"Outerwear","href":"/shop"}
    ]},
    {"title":"Shoes","links":[
      {"title":"Heels","href":"/shop"},
      {"title":"Flats","href":"/shop"},
      {"title":"Boots","href":"/shop"},
      {"title":"Sneakers","href":"/shop"}
    ]},
    {"title":"Bags","links":[
      {"title":"Tote Bags","href":"/shop"},
      {"title":"Crossbody","href":"/shop"},
      {"title":"Clutches","href":"/shop"},
      {"title":"Backpacks","href":"/shop"}
    ]}
  ]'::jsonb,
  '[
    {"title":"Spring Edit","description":"Fresh silhouettes for the season","href":"/collections","image":"https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=400&h=500&fit=crop","badge":"New"}
  ]'::jsonb,
  1,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Women');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Men',
  'mega',
  NULL,
  NULL,
  '{"label":"Shop all men''s","href":"/shop"}'::jsonb,
  '[
    {"title":"Clothing","links":[
      {"title":"Shirts","href":"/shop"},
      {"title":"Trousers","href":"/shop"},
      {"title":"Jackets","href":"/shop"},
      {"title":"Activewear","href":"/shop"}
    ]},
    {"title":"Shoes","links":[
      {"title":"Loafers","href":"/shop"},
      {"title":"Boots","href":"/shop"},
      {"title":"Sneakers","href":"/shop"},
      {"title":"Formal","href":"/shop"}
    ]},
    {"title":"Watches","links":[
      {"title":"Dress Watches","href":"/shop"},
      {"title":"Sport Watches","href":"/shop"},
      {"title":"Chronographs","href":"/shop"},
      {"title":"Limited Editions","href":"/shop?sortBy=newest"}
    ]}
  ]'::jsonb,
  '[
    {"title":"Tailored Essentials","description":"Refined staples for every day","href":"/shop","image":"https://images.unsplash.com/photo-1617137968427-85924c800a22?w=400&h=500&fit=crop","badge":"Featured"}
  ]'::jsonb,
  2,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Men');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Accessories',
  'mega',
  NULL,
  NULL,
  '{"label":"Shop all accessories","href":"/shop"}'::jsonb,
  '[
    {"title":"Jewelry","links":[
      {"title":"Necklaces","href":"/shop"},
      {"title":"Earrings","href":"/shop"},
      {"title":"Rings","href":"/shop"},
      {"title":"Bracelets","href":"/shop"}
    ]},
    {"title":"Watches","links":[
      {"title":"Automatic","href":"/shop"},
      {"title":"Chronograph","href":"/shop"},
      {"title":"Dress Watches","href":"/shop"},
      {"title":"Smart Watches","href":"/shop"}
    ]},
    {"title":"Eyewear","links":[
      {"title":"Sunglasses","href":"/shop"},
      {"title":"Optical Frames","href":"/shop"},
      {"title":"Blue Light","href":"/shop"},
      {"title":"Limited Drops","href":"/shop?sortBy=newest"}
    ]}
  ]'::jsonb,
  '[
    {"title":"Iconic Timepieces","description":"Curated watches from top sellers","href":"/shop","image":"https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=400&h=500&fit=crop","badge":"Trending"}
  ]'::jsonb,
  3,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Accessories');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'New Arrivals',
  'link',
  '/shop?sortBy=newest',
  'New',
  NULL,
  NULL,
  NULL,
  4,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'New Arrivals');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Sale',
  'link',
  '/shop?showOnlySale=true',
  'Sale',
  NULL,
  NULL,
  NULL,
  5,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Sale');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Collections',
  'link',
  '/collections',
  NULL,
  NULL,
  NULL,
  NULL,
  6,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Collections');

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Gift Cards',
  'mega',
  NULL,
  NULL,
  '{"label":"Shop gift cards","href":"/gift-cards"}'::jsonb,
  '[
    {"title":"Gifting","links":[
      {"title":"Buy a gift card","href":"/gift-cards"},
      {"title":"Gift finder","href":"/gift-cards/finder"}
    ]}
  ]'::jsonb,
  NULL,
  7,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Gift Cards');

-- Consolidate gift nav on existing databases (dropdown + remove duplicate top-level item)
UPDATE nav_menus
SET
  type = 'mega',
  href = NULL,
  badge = NULL,
  view_all = '{"label":"Shop gift cards","href":"/gift-cards"}'::jsonb,
  columns = '[
    {"title":"Gifting","links":[
      {"title":"Buy a gift card","href":"/gift-cards"},
      {"title":"Gift finder","href":"/gift-cards/finder"}
    ]}
  ]'::jsonb,
  featured = NULL,
  updated_at = NOW()
WHERE label = 'Gift Cards';

DELETE FROM nav_menus WHERE label = 'Gift Finder';

-- Roles & permissions (idempotent)
INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Administrator', 'admin', 'Full platform access', true, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'admin');

INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Customer', 'user', 'Standard storefront account', true, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'user');

INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Moderator', 'moderator', 'Limited admin dashboard access', true, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'moderator');

INSERT INTO permissions (key, module, description, created_at, updated_at)
SELECT v.key, v.module, v.description, NOW(), NOW()
FROM (VALUES
  ('users.read', 'users', 'View user accounts'),
  ('users.write', 'users', 'Manage user accounts and roles'),
  ('orders.read', 'orders', 'View orders'),
  ('orders.write', 'orders', 'Update order status'),
  ('products.read', 'products', 'View products'),
  ('products.write', 'products', 'Create and edit products'),
  ('menus.read', 'menus', 'View admin and site menus'),
  ('menus.write', 'menus', 'Manage admin and site menus'),
  ('roles.read', 'roles', 'View roles and permissions'),
  ('roles.write', 'roles', 'Manage roles and permissions'),
  ('teams.read', 'teams', 'View teams and members'),
  ('teams.write', 'teams', 'Manage teams and membership'),
  ('settings.read', 'settings', 'View platform settings'),
  ('settings.write', 'settings', 'Update platform settings')
) AS v(key, module, description)
WHERE NOT EXISTS (SELECT 1 FROM permissions p WHERE p.key = v.key);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'admin'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('orders.read', 'products.read', 'users.read')
WHERE r.slug = 'moderator'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Content Manager', 'content-manager', 'Manage catalog, collections, and site navigation', false, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'content-manager');

INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Support Agent', 'support', 'Handle customer accounts and order lookups', false, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'support');

INSERT INTO roles (name, slug, description, is_system, created_at, updated_at)
SELECT 'Catalog Manager', 'catalog-manager', 'Products, categories, and merchandising', false, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE slug = 'catalog-manager');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN (
  'products.read', 'products.write', 'menus.read', 'menus.write', 'orders.read'
)
WHERE r.slug = 'content-manager'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('users.read', 'orders.read')
WHERE r.slug = 'support'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('products.read', 'products.write', 'menus.read')
WHERE r.slug = 'catalog-manager'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

-- Keep catalog search index in sync for demo rows inserted without search_document.
UPDATE products
SET search_document = trim(both FROM concat_ws(' ', name, description, sku, barcode, slug, array_to_string(tags, ' ')))
WHERE (search_document IS NULL OR btrim(search_document) = '')
  AND slug LIKE 'pdp-demo%';
