-- Check existing products
SELECT id, name FROM products LIMIT 10;

-- Insert/update compare list for user 4 (using product IDs 1,2,3 as example)
INSERT INTO  compare_lists (user_id, product_ids)
VALUES (4, ARRAY[1, 2, 3])
ON CONFLICT (user_id) DO UPDATE SET product_ids = EXCLUDED.product_ids;