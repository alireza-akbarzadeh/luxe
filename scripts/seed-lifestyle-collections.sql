-- Lifestyle-themed curated collections (Task 007). Run via: make seed-dev
-- Idempotent: ON CONFLICT (slug) updates theme and merchandising fields.

INSERT INTO collections (
    slug, eyebrow, title, description, href, image_url, cta_label, sort_order, status, theme, created_at, updated_at
) VALUES
    (
        'minimal-workspace',
        'Lifestyle',
        'Minimal Workspace',
        'Clean lines, quiet materials, and pieces that keep your desk calm — watches, organizers, and lighting that earn their place.',
        '/shop?search=watch',
        'https://images.unsplash.com/photo-1497366216548-37526070297c?w=1200',
        'Shop the edit',
        10,
        'active',
        'lifestyle',
        NOW(),
        NOW()
    ),
    (
        'travel-essentials',
        'Lifestyle',
        'Travel Essentials',
        'Compact carry, durable finishes, and timeless accessories built for terminals, hotels, and time zones.',
        '/shop?search=leather',
        'https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=1200',
        'Pack smarter',
        11,
        'active',
        'lifestyle',
        NOW(),
        NOW()
    ),
    (
        'first-apartment',
        'Lifestyle',
        'First Apartment',
        'Foundational pieces for a first place of your own — elevated everyday objects that make a small space feel intentional.',
        '/shop',
        'https://images.unsplash.com/photo-1556909114-f6e7ad7d4046?w=1200',
        'Start here',
        12,
        'active',
        'lifestyle',
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
    theme = EXCLUDED.theme,
    updated_at = NOW();
