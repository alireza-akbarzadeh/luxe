-- +goose Up
-- +goose StatementBegin
-- Rich watch metadata for PDP feature grid (English labels, Digikala-style density).

DO $$
DECLARE
    prod_id INT;
BEGIN
    SELECT id INTO prod_id FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL LIMIT 1;
    IF prod_id IS NULL THEN
        RETURN;
    END IF;

    UPDATE product_attributes
    SET values = ARRAY['Silver', 'Gold'], updated_at = NOW()
    WHERE product_id = prod_id AND lower(name) = 'color' AND deleted_at IS NULL;

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, 'color', ARRAY['Silver', 'Gold'], NOW(), NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes
        WHERE product_id = prod_id AND lower(name) = 'color' AND deleted_at IS NULL
    );

    UPDATE product_attributes
    SET values = ARRAY['40mm', '42mm', '44mm'], updated_at = NOW()
    WHERE product_id = prod_id AND lower(name) = 'size' AND deleted_at IS NULL;

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, 'size', ARRAY['40mm', '42mm', '44mm'], NOW(), NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes
        WHERE product_id = prod_id AND lower(name) = 'size' AND deleted_at IS NULL
    );

    INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
    SELECT prod_id, v.name, v.values, NOW(), NOW()
    FROM (VALUES
        ('user_style',           ARRAY['Sport', 'Daily wear', 'Formal']),
        ('dial_shape',           ARRAY['Sunburst round']),
        ('crystal_material',     ARRAY['Sapphire anti-reflective']),
        ('case_material',        ARRAY['Stainless steel']),
        ('bezel_material',       ARRAY['Fixed tachymeter bezel']),
        ('clasp_type',           ARRAY['Deployment buckle']),
        ('strap_material',       ARRAY['Leather', 'Steel bracelet', 'Rubber']),
        ('power_source',         ARRAY['Automatic (self-winding)']),
        ('specialized_features', ARRAY['Chronograph', '24-hour display', 'Date window', 'Luminous hands']),
        ('movement',             ARRAY['Swiss automatic']),
        ('water_resistance',     ARRAY['100m']),
        ('power_reserve',        ARRAY['42 hours']),
        ('warranty',             ARRAY['2 years international']),
        ('origin',               ARRAY['Swiss Made'])
    ) AS v(name, values)
    WHERE NOT EXISTS (
        SELECT 1 FROM product_attributes pa
        WHERE pa.product_id = prod_id AND lower(pa.name) = lower(v.name) AND pa.deleted_at IS NULL
    );

    UPDATE products
    SET
        colors = '["Silver","Gold"]'::jsonb,
        sizes  = '["40mm","42mm","44mm"]'::jsonb,
        updated_at = NOW()
    WHERE id = prod_id;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM product_attributes
WHERE product_id IN (SELECT id FROM products WHERE slug = 'pdp-demo-luxe-watch')
  AND lower(name) IN (
    'user_style', 'dial_shape', 'crystal_material', 'case_material', 'bezel_material',
    'clasp_type', 'strap_material', 'power_source', 'specialized_features',
    'movement', 'water_resistance', 'power_reserve', 'warranty', 'origin'
  );
-- +goose StatementEnd
