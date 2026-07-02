-- Homepage merchandising demo data: published collections + active flash deals.
-- Run after seed-catalog.sql. Idempotent.

-- Active flash deal (powers PromoSection on the storefront home page)
INSERT INTO flash_deals (product_id, ends_at, quantity_limit, sort_order, status, created_at, updated_at)
SELECT
    p.id,
    NOW() + INTERVAL '14 days',
    50,
    0,
    'active',
    NOW(),
    NOW()
FROM products p
WHERE p.deleted_at IS NULL
  AND p.status = 'active'
  AND p.images IS NOT NULL
  AND p.images::text <> '[]'
  AND NOT EXISTS (
    SELECT 1 FROM flash_deals fd
    WHERE fd.product_id = p.id AND fd.status = 'active' AND fd.ends_at > NOW()
  )
ORDER BY p.compare_at_price DESC NULLS LAST, p.id ASC
LIMIT 1;

-- Published collections (powers CollectionBanner on the storefront home page)
INSERT INTO collections (
    slug, eyebrow, title, description, href, image_url, cta_label, sort_order, status, created_at, updated_at
)
VALUES
    (
        'fall-edit',
        'New season',
        'The Fall Edit',
        'Layered tailoring, rich textures, and timeless silhouettes for the season ahead.',
        '/shop?sortBy=newest',
        'https://images.unsplash.com/photo-1483985988355-763728e3685b?w=1200',
        'Shop the edit',
        0,
        'published',
        NOW(),
        NOW()
    ),
    (
        'evening-essentials',
        'Occasion wear',
        'Evening Essentials',
        'Silk, satin, and statement pieces curated for after dark.',
        '/shop',
        'https://images.unsplash.com/photo-1566174053879-31528523f8ae?w=1200',
        'Explore collection',
        1,
        'published',
        NOW(),
        NOW()
    ),
    (
        'modern-minimal',
        'Capsule',
        'Modern Minimal',
        'Clean lines and elevated basics for a refined everyday wardrobe.',
        '/shop',
        'https://images.unsplash.com/photo-1445205170230-053b83016050?w=1200',
        'View collection',
        2,
        'published',
        NOW(),
        NOW()
    ),
    (
        'accessories-edit',
        'Finishing touches',
        'The Accessories Edit',
        'Watches, leather goods, and jewelry to complete every look.',
        '/shop',
        'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=1200',
        'Shop accessories',
        3,
        'published',
        NOW(),
        NOW()
    )
ON CONFLICT (slug) DO UPDATE SET
    eyebrow = EXCLUDED.eyebrow,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    href = EXCLUDED.href,
    image_url = EXCLUDED.image_url,
    cta_label = EXCLUDED.cta_label,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW();
