-- +goose Up
ALTER TABLE products
    ADD COLUMN brand_id INT,
    ADD CONSTRAINT fk_products_brand
        FOREIGN KEY (brand_id) REFERENCES brands(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE products
    DROP CONSTRAINT IF EXISTS fk_products_brand,
    DROP COLUMN IF EXISTS brand_id;