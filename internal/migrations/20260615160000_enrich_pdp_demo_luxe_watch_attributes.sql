-- +goose Up
-- +goose StatementBegin
-- Rich attributes + gallery images for PDP demo watch (by slug, not hard-coded id).

DO $$
DECLARE
    prod_id INT;
BEGIN
    SELECT id INTO prod_id FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL LIMIT 1;
    IF prod_id IS NULL THEN
        RETURN;
    END IF;

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, 'color', ARRAY['#C0C0C0', '#D4AF37', '#1A1A1A', '#B76E79'], NOW(), NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes
        WHERE product_id = prod_id AND lower(name) = 'color' AND deleted_at IS NULL
    );

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, 'size', ARRAY['38mm', '40mm', '42mm', '44mm'], NOW(), NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes
        WHERE product_id = prod_id AND lower(name) = 'size' AND deleted_at IS NULL
    );

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, v.name, v.values, NOW(), NOW()
    FROM (VALUES
        ('strap',         ARRAY['Leather', 'Steel bracelet', 'NATO', 'Rubber']),
        ('clasp',         ARRAY['Deployment buckle', 'Pin buckle']),
        ('dimensions',    ARRAY['Case: 42 × 11.5 mm', 'Lug width: 22 mm', 'Weight: 152 g']),
        ('crystal',       ARRAY['Sapphire anti-reflective']),
        ('power_reserve', ARRAY['42 hours']),
        ('warranty',      ARRAY['2 years international']),
        ('origin',        ARRAY['Swiss Made']),
        ('handmade',      ARRAY['false'])
    ) AS v(name, values)
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes pa
        WHERE pa.product_id = prod_id AND lower(pa.name) = lower(v.name) AND pa.deleted_at IS NULL
    );

    UPDATE products
    SET
        colors = '["Silver","Gold","Midnight Black","Rose Gold"]'::jsonb,
        sizes  = '["38mm","40mm","42mm","44mm"]'::jsonb,
        images = ARRAY[
            'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=800',
            'https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=800',
            'https://images.unsplash.com/photo-1587836374828-4db968944bda?w=800',
            'https://images.unsplash.com/photo-1614162692292-7c008625e467?w=800',
            'https://images.unsplash.com/photo-1547996160-20dfa5d0dfd1?w=800',
            'https://images.unsplash.com/photo-1611658890348-2b2546a56333?w=800'
        ],
        updated_at = NOW()
    WHERE id = prod_id;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM product_attributes
WHERE product_id IN (SELECT id FROM products WHERE slug = 'pdp-demo-luxe-watch')
  AND lower(name) IN (
    'color', 'size', 'strap', 'clasp', 'dimensions', 'crystal',
    'power_reserve', 'warranty', 'origin', 'handmade'
  );
-- +goose StatementEnd
