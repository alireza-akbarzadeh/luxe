-- Discover mega menu for storefront navbar (Brands, Collections, Stores, etc.)
-- Flat links live in columns JSON (not parent/child nav rows).
-- Safe to re-run: inserts when missing, then upserts column content.

INSERT INTO nav_menus (label, type, href, badge, view_all, columns, featured, "order", created_at, updated_at)
SELECT
  'Discover',
  'mega',
  NULL,
  NULL,
  '{"label":"Explore Luxe","href":"/shop"}'::jsonb,
  '[
    {"title":"Shop","links":[
      {"title":"Shop all","href":"/shop"},
      {"title":"New arrivals","href":"/shop?sortBy=newest&showOnlyNew=true"},
      {"title":"Best sellers","href":"/shop?sortBy=rating_desc"},
      {"title":"Sale","href":"/shop?showOnlySale=true"}
    ]},
    {"title":"Discover","links":[
      {"title":"Brands","href":"/brands"},
      {"title":"Collections","href":"/collections"},
      {"title":"Stores","href":"/store"},
      {"title":"Gift cards","href":"/gift-cards"}
    ]},
    {"title":"More","links":[
      {"title":"Luxe Plus","href":"/plus/landing"},
      {"title":"Sell on Luxe","href":"/vendor"},
      {"title":"Support","href":"/support"},
      {"title":"Lifestyle","href":"/lifestyle"}
    ]}
  ]'::jsonb,
  '[
    {"title":"Shop by brand","description":"200+ premium maisons on Luxe","href":"/brands","image":"https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&h=500&fit=crop","badge":"New"},
    {"title":"Discover stores","description":"Verified sellers and boutiques","href":"/store","image":"https://images.unsplash.com/photo-1441984904996-e0b6ba687e04?w=400&h=500&fit=crop","badge":"Featured"}
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
  view_all = '{"label":"Explore Luxe","href":"/shop"}'::jsonb,
  columns = '[
    {"title":"Shop","links":[
      {"title":"Shop all","href":"/shop"},
      {"title":"New arrivals","href":"/shop?sortBy=newest&showOnlyNew=true"},
      {"title":"Best sellers","href":"/shop?sortBy=rating_desc"},
      {"title":"Sale","href":"/shop?showOnlySale=true"}
    ]},
    {"title":"Discover","links":[
      {"title":"Brands","href":"/brands"},
      {"title":"Collections","href":"/collections"},
      {"title":"Stores","href":"/store"},
      {"title":"Gift cards","href":"/gift-cards"}
    ]},
    {"title":"More","links":[
      {"title":"Luxe Plus","href":"/plus/landing"},
      {"title":"Sell on Luxe","href":"/vendor"},
      {"title":"Support","href":"/support"},
      {"title":"Lifestyle","href":"/lifestyle"}
    ]}
  ]'::jsonb,
  featured = '[
    {"title":"Shop by brand","description":"200+ premium maisons on Luxe","href":"/brands","image":"https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&h=500&fit=crop","badge":"New"},
    {"title":"Discover stores","description":"Verified sellers and boutiques","href":"/store","image":"https://images.unsplash.com/photo-1441984904996-e0b6ba687e04?w=400&h=500&fit=crop","badge":"Featured"}
  ]'::jsonb,
  "order" = 0,
  updated_at = NOW()
WHERE label = 'Discover';

SELECT id, label, type, "order", columns
FROM nav_menus
WHERE label = 'Discover';
