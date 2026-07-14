-- Discover mega menu (flat links live in columns, not parent/child nav rows)
INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Discover',
  'mega',
  NULL,
  NULL,
  '{"label":"Browse all","href":"/products"}'::jsonb,
  '[
    {"title":"Explore","links":[
      {"title":"Explore","href":"/lifestyle"},
      {"title":"Shop","href":"/shop"},
      {"title":"Collections","href":"/collections"},
      {"title":"Browse","href":"/products"}
    ]},
    {"title":"Featured","links":[
      {"title":"Featured","href":"/products/best-sellers"},
      {"title":"Curated","href":"/public-collections"},
      {"title":"Luxury Picks","href":"/plus/landing"},
      {"title":"New & Featured","href":"/shop?sortBy=newest"}
    ]},
    {"title":"Marketplace","links":[
      {"title":"Marketplace","href":"/store"}
    ]}
  ]'::jsonb,
  '[
    {"title":"New & Featured","description":"Fresh arrivals and editor picks","href":"/shop?sortBy=newest","image":"https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&h=500&fit=crop","badge":"New"}
  ]'::jsonb,
  0,
  NOW(),
  NOW()
WHERE NOT EXISTS (SELECT 1 FROM nav_menus WHERE label = 'Discover');

UPDATE nav_menus
SET
  type = 'mega',
  href = NULL,
  badge = NULL,
  view_all = '{"label":"Browse all","href":"/products"}'::jsonb,
  columns = '[
    {"title":"Explore","links":[
      {"title":"Explore","href":"/lifestyle"},
      {"title":"Shop","href":"/shop"},
      {"title":"Collections","href":"/collections"},
      {"title":"Browse","href":"/products"}
    ]},
    {"title":"Featured","links":[
      {"title":"Featured","href":"/products/best-sellers"},
      {"title":"Curated","href":"/public-collections"},
      {"title":"Luxury Picks","href":"/plus/landing"},
      {"title":"New & Featured","href":"/shop?sortBy=newest"}
    ]},
    {"title":"Marketplace","links":[
      {"title":"Marketplace","href":"/store"}
    ]}
  ]'::jsonb,
  featured = '[
    {"title":"New & Featured","description":"Fresh arrivals and editor picks","href":"/shop?sortBy=newest","image":"https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&h=500&fit=crop","badge":"New"}
  ]'::jsonb,
  "order" = 0,
  updated_at = NOW()
WHERE label = 'Discover';

SELECT id, label, type, "order", columns
FROM nav_menus
WHERE label = 'Discover';
