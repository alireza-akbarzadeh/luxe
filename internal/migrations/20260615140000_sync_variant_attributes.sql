-- +goose Up
-- +goose StatementBegin
-- Mirror legacy colors/sizes JSON columns into product_attributes for PDP variant pickers.

INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
SELECT p.id, 'color', ARRAY(SELECT jsonb_array_elements_text(p.colors)), NOW(), NOW()
FROM products p
WHERE jsonb_array_length(COALESCE(p.colors, '[]'::jsonb)) > 0
  AND NOT EXISTS (
    SELECT 1 FROM product_attributes pa
    WHERE pa.product_id = p.id
      AND pa.deleted_at IS NULL
      AND lower(pa.name) IN ('color', 'colors', 'colour', 'colours')
  );

INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
SELECT p.id, 'size', ARRAY(SELECT jsonb_array_elements_text(p.sizes)), NOW(), NOW()
FROM products p
WHERE jsonb_array_length(COALESCE(p.sizes, '[]'::jsonb)) > 0
  AND NOT EXISTS (
    SELECT 1 FROM product_attributes pa
    WHERE pa.product_id = p.id
      AND pa.deleted_at IS NULL
      AND lower(pa.name) IN ('size', 'sizes')
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM product_attributes
WHERE lower(name) IN ('color', 'colors', 'colour', 'colours', 'size', 'sizes')
  AND product_id IN (
    SELECT id FROM products
    WHERE jsonb_array_length(COALESCE(colors, '[]'::jsonb)) > 0
       OR jsonb_array_length(COALESCE(sizes, '[]'::jsonb)) > 0
  );
-- +goose StatementEnd
