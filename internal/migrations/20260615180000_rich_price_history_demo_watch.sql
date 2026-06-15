-- +goose Up
-- +goose StatementBegin
-- Replace flat price history with a 90-day trend (sales dips + recovery) for demo watch.

DO $$
DECLARE
    prod_id INT;
BEGIN
    SELECT id INTO prod_id FROM products WHERE slug = 'pdp-demo-luxe-watch' AND deleted_at IS NULL LIMIT 1;
    IF prod_id IS NULL THEN
        RETURN;
    END IF;

    DELETE FROM product_price_history WHERE product_id = prod_id;

    INSERT INTO product_price_history (product_id, price, compare_at_price, recorded_at)
    SELECT prod_id, v.price, v.compare_val, NOW() - (v.days_ago || ' days')::interval
    FROM (VALUES
        (88, 1599.00::decimal, 1799.00::decimal),
        (84, 1579.00::decimal, 1779.00::decimal),
        (80, 1549.00::decimal, 1749.00::decimal),
        (76, 1519.00::decimal, 1719.00::decimal),
        (72, 1489.00::decimal, 1699.00::decimal),
        (68, 1449.00::decimal, 1649.00::decimal),
        (64, 1399.00::decimal, 1599.00::decimal),
        (60, 1349.00::decimal, 1549.00::decimal),
        (56, 1279.00::decimal, 1499.00::decimal),
        (52, 1229.00::decimal, 1449.00::decimal),
        (48, 1199.00::decimal, 1399.00::decimal),
        (44, 1219.00::decimal, 1419.00::decimal),
        (40, 1249.00::decimal, 1449.00::decimal),
        (36, 1279.00::decimal, 1479.00::decimal),
        (32, 1299.00::decimal, 1499.00::decimal),
        (28, 1319.00::decimal, 1519.00::decimal),
        (24, 1339.00::decimal, 1539.00::decimal),
        (20, 1329.00::decimal, 1529.00::decimal),
        (14, 1309.00::decimal, 1509.00::decimal),
        (10, 1299.00::decimal, 1499.00::decimal),
        (6,  1289.00::decimal, 1489.00::decimal),
        (3,  1299.00::decimal, 1499.00::decimal),
        (0,  1299.00::decimal, 1499.00::decimal)
    ) AS v(days_ago, price, compare_val);
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM product_price_history
WHERE product_id IN (SELECT id FROM products WHERE slug = 'pdp-demo-luxe-watch');
-- +goose StatementEnd
