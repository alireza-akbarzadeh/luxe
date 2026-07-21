-- Homepage merchandising demo data: hero slides, flash promo band, flash deals, collections.
-- Run after seed-catalog.sql. Idempotent.

-- ── Flash-deals promo copy (powers PromoSection headline + countdown) ───────────
INSERT INTO homepage_sections (
    section_key, title, href, image_url, sort_order, status, filters, created_at, updated_at
)
VALUES (
    'flash-deals-promo',
    '18% off your first order',
    '/shop',
    '',
    0,
    'published',
    jsonb_build_object(
        'badge', 'Limited time',
        'description', 'Unlock 18% off your first order — plus early access to private sales, new drops, and member-only styling sessions. Use code WELCOME30 at checkout.',
        'cta_label', 'Shop the sale',
        'ends_at', to_char((NOW() + INTERVAL '5 days') AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
        'theme', 'dark'
    ),
    NOW(),
    NOW()
)
ON CONFLICT (section_key) DO UPDATE SET
    title = EXCLUDED.title,
    href = EXCLUDED.href,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    filters = EXCLUDED.filters,
    updated_at = NOW();

-- ── Second marketing band (section_key must start with marketing-band-) ───────
INSERT INTO homepage_sections (
    section_key, title, href, image_url, sort_order, status, filters, created_at, updated_at
)
VALUES (
    'marketing-band-new-members',
    'Members save 15% on tailoring',
    '/shop?category=tailoring',
    '',
    1,
    'published',
    jsonb_build_object(
        'badge', 'New members',
        'description', 'Join Luxe Plus for exclusive pricing on made-to-measure pieces and complimentary alterations.',
        'cta_label', 'Explore tailoring',
        'ends_at', to_char((NOW() + INTERVAL '7 days') AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
        'theme', 'light',
        'flash_deal_ids', (
            SELECT COALESCE(jsonb_agg(fd.id ORDER BY fd.sort_order), '[]'::jsonb)
            FROM flash_deals fd
            INNER JOIN products p ON p.id = fd.product_id
            WHERE fd.status = 'active'
              AND fd.ends_at > NOW()
              AND p.slug IN ('heritage-chronograph-42', 'stellar-automatic-38', 'solstice-gold-vermeil-hoops', 'sculptural-leather-pump')
        )
    ),
    NOW(),
    NOW()
)
ON CONFLICT (section_key) DO UPDATE SET
    title = EXCLUDED.title,
    href = EXCLUDED.href,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    filters = EXCLUDED.filters,
    updated_at = NOW();

-- ── Hero carousel slides (section_key must start with hero-) ──────────────────
INSERT INTO homepage_sections (
    section_key, title, href, image_url, sort_order, status, filters, created_at, updated_at
)
VALUES
    (
        'hero-slide-1',
        'Up to 40% off selected pieces',
        '/shop',
        'https://images.unsplash.com/photo-1490481651871-ab68de25d43d?w=1200&h=900&fit=crop',
        0,
        'published',
        '{"eyebrow":"Seasonal edit","description":"Curated luxury for everyday life"}'::jsonb,
        NOW(),
        NOW()
    ),
    (
        'hero-slide-2',
        'The Fall Edit is here',
        '/shop?sortBy=newest',
        'https://images.unsplash.com/photo-1483985988355-763728e3685b?w=1200&h=900&fit=crop',
        1,
        'published',
        '{"eyebrow":"New arrivals","description":"Layered tailoring and rich textures"}'::jsonb,
        NOW(),
        NOW()
    ),
    (
        'hero-slide-3',
        'Evening essentials',
        '/collections',
        'https://images.unsplash.com/photo-1566174053879-31528523f8ae?w=1200&h=900&fit=crop',
        2,
        'published',
        '{"eyebrow":"Occasion wear","description":"Silk, satin, and statement pieces"}'::jsonb,
        NOW(),
        NOW()
    )
ON CONFLICT (section_key) DO UPDATE SET
    title = EXCLUDED.title,
    href = EXCLUDED.href,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    filters = EXCLUDED.filters,
    updated_at = NOW();

-- ── Seasonal pick banners (shown in SeasonalPicksSection) ─────────────────────
INSERT INTO homepage_sections (
    section_key, title, href, image_url, sort_order, status, filters, created_at, updated_at
)
VALUES
    (
        'seasonal-gifts',
        'Gift-worthy finds',
        '/shop',
        'https://images.unsplash.com/photo-1513883049090-d0b7439799a8?w=1200',
        10,
        'published',
        '{}'::jsonb,
        NOW(),
        NOW()
    ),
    (
        'seasonal-outerwear',
        'Outerwear spotlight',
        '/shop',
        'https://images.unsplash.com/photo-1548126032-077a2e8e9e3b?w=1200',
        11,
        'published',
        '{}'::jsonb,
        NOW(),
        NOW()
    )
ON CONFLICT (section_key) DO UPDATE SET
    title = EXCLUDED.title,
    href = EXCLUDED.href,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    filters = EXCLUDED.filters,
    updated_at = NOW();

-- ── Flash deals product rail (powers PromoSection product cards) ──────────────
INSERT INTO flash_deals (product_id, title, ends_at, quantity_limit, sort_order, status, created_at, updated_at)
SELECT
    p.id,
    '',
    NOW() + INTERVAL '5 days',
    50,
    ord.sort_order,
    'active',
    NOW(),
    NOW()
FROM (
    VALUES
        ('arielle-silk-midi-dress', 0),
        ('soft-leather-biker-jacket', 1),
        ('heritage-chronograph-42', 2),
        ('cloud-cashmere-crew', 3),
        ('stellar-automatic-38', 4),
        ('sculptural-leather-pump', 5),
        ('solstice-gold-vermeil-hoops', 6),
        ('city-rain-trench', 7)
) AS ord(slug, sort_order)
INNER JOIN products p ON p.slug = ord.slug AND p.deleted_at IS NULL
WHERE p.status = 'active'
  AND NOT EXISTS (
    SELECT 1
    FROM flash_deals fd
    WHERE fd.product_id = p.id
      AND fd.status = 'active'
      AND fd.ends_at > NOW()
  );

-- Refresh ends_at on existing active demo deals so countdown stays current
UPDATE flash_deals fd
SET
    ends_at = NOW() + INTERVAL '5 days',
    sort_order = ord.sort_order,
    updated_at = NOW()
FROM products p
INNER JOIN (
    VALUES
        ('arielle-silk-midi-dress', 0),
        ('soft-leather-biker-jacket', 1),
        ('heritage-chronograph-42', 2),
        ('cloud-cashmere-crew', 3),
        ('stellar-automatic-38', 4),
        ('sculptural-leather-pump', 5),
        ('solstice-gold-vermeil-hoops', 6),
        ('city-rain-trench', 7)
) AS ord(slug, sort_order) ON p.slug = ord.slug
WHERE fd.product_id = p.id
  AND fd.status = 'active'
  AND fd.ends_at > NOW();

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
        'active',
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
        'active',
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
        'active',
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
        'active',
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
