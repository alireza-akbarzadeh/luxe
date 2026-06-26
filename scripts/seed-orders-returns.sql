-- Dev/staging demo orders for vendor panel testing.
-- Run: make seed-orders-returns (after seed-dev + seed-catalog)
-- Idempotent: skips when demo vendor orders already exist.

DO $$
DECLARE
    vendor_user_id BIGINT;
    buyer_user_id BIGINT;
    store_luxe_id BIGINT;
    prod_watch_id BIGINT;
    prod_oos_id BIGINT;
    ord_id BIGINT;
    pay_id BIGINT;
    ship_id BIGINT;
BEGIN
    SELECT id INTO vendor_user_id FROM users ORDER BY id ASC LIMIT 1;
    SELECT id INTO buyer_user_id FROM users ORDER BY id ASC OFFSET 1 LIMIT 1;
    IF buyer_user_id IS NULL THEN
        buyer_user_id := vendor_user_id;
    END IF;

    SELECT id INTO store_luxe_id FROM stores WHERE slug = 'luxe-atelier';
    SELECT id INTO prod_watch_id FROM products WHERE slug = 'pdp-demo-luxe-watch';
    SELECT id INTO prod_oos_id FROM products WHERE slug = 'pdp-demo-watch-sold-out';

    IF store_luxe_id IS NULL OR prod_watch_id IS NULL OR vendor_user_id IS NULL THEN
        RAISE NOTICE 'seed-orders-returns: missing store/product/user — run seed-dev first';
        RETURN;
    END IF;

    -- Assign Luxe Atelier to the first user so vendor panel can manage it
    UPDATE stores SET user_id = vendor_user_id WHERE id = store_luxe_id AND user_id IS NULL;

    IF EXISTS (SELECT 1 FROM orders WHERE order_number LIKE 'VND-DEMO-%') THEN
        RAISE NOTICE 'seed-orders-returns: demo orders already present — skipping';
        RETURN;
    END IF;

    -- Order 1: pending
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1001', 'pending', 1299.00, 'USD', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'credit_card', 'pending', 'txn_vnd_1001', NOW(), NOW())
    RETURNING id INTO pay_id;
    UPDATE orders SET payment_id = pay_id WHERE id = ord_id;

    -- Order 2: paid
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1002', 'paid', 2598.00, 'USD', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 2, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 2598.00, 'USD', 'credit_card', 'succeeded', 'txn_vnd_1002', NOW(), NOW())
    RETURNING id INTO pay_id;
    UPDATE orders SET payment_id = pay_id WHERE id = ord_id;

    -- Order 3: shipped with tracking
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1003', 'shipped', 1299.00, 'USD', NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'paypal', 'succeeded', 'txn_vnd_1003', NOW(), NOW())
    RETURNING id INTO pay_id;
    INSERT INTO shipments (
        order_id, user_id, carrier, tracking_number, status, shipped_at,
        address_line1, city, state, postal_code, country, created_at, updated_at
    ) VALUES (
        ord_id, buyer_user_id, 'DHL Express', 'DHL-VND-7845123', 'shipped', NOW() - INTERVAL '1 day',
        '123 Fashion Ave', 'Tehran', 'Tehran', '1234567890', 'IR', NOW(), NOW()
    ) RETURNING id INTO ship_id;
    UPDATE orders SET payment_id = pay_id, shipment_id = ship_id WHERE id = ord_id;

    -- Order 4: delivered
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1004', 'delivered', 1299.00, 'USD', NOW() - INTERVAL '10 days', NOW() - INTERVAL '2 days')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'credit_card', 'succeeded', 'txn_vnd_1004', NOW(), NOW())
    RETURNING id INTO pay_id;
    INSERT INTO shipments (
        order_id, user_id, carrier, tracking_number, status, shipped_at, delivered_at,
        address_line1, city, state, postal_code, country, created_at, updated_at
    ) VALUES (
        ord_id, buyer_user_id, 'FedEx', 'FX-VND-991234', 'delivered', NOW() - INTERVAL '7 days', NOW() - INTERVAL '2 days',
        '45 Luxury Blvd', 'Isfahan', 'Isfahan', '9876543210', 'IR', NOW(), NOW()
    ) RETURNING id INTO ship_id;
    UPDATE orders SET payment_id = pay_id, shipment_id = ship_id WHERE id = ord_id;

    -- Order 5: cancelled
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1005', 'cancelled', 1299.00, 'USD', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'credit_card', 'failed', 'txn_vnd_1005', NOW(), NOW())
    RETURNING id INTO pay_id;
    UPDATE orders SET payment_id = pay_id WHERE id = ord_id;

    -- Order 6: refunded
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1006', 'refunded', 1299.00, 'USD', NOW() - INTERVAL '14 days', NOW() - INTERVAL '3 days')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'credit_card', 'refunded', 'txn_vnd_1006', NOW(), NOW())
    RETURNING id INTO pay_id;
    UPDATE orders SET payment_id = pay_id WHERE id = ord_id;

    -- Order 7: paid today
    INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
    VALUES (buyer_user_id, 'VND-DEMO-1007', 'paid', 1299.00, 'USD', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes')
    RETURNING id INTO ord_id;
    INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
    VALUES (ord_id, prod_watch_id, 1, 1299.00, NOW(), NOW());
    INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
    VALUES (ord_id, buyer_user_id, 1299.00, 'USD', 'debit_card', 'succeeded', 'txn_vnd_1007', NOW(), NOW())
    RETURNING id INTO pay_id;
    UPDATE orders SET payment_id = pay_id WHERE id = ord_id;

    -- Order 8: paid with OOS product if available
    IF prod_oos_id IS NOT NULL THEN
        INSERT INTO orders (user_id, order_number, status, total_amount, currency, created_at, updated_at)
        VALUES (buyer_user_id, 'VND-DEMO-1008', 'paid', 899.00, 'USD', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours')
        RETURNING id INTO ord_id;
        INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
        VALUES (ord_id, prod_oos_id, 1, 899.00, NOW(), NOW());
        INSERT INTO payments (order_id, user_id, amount, currency, method, status, transaction_id, created_at, updated_at)
        VALUES (ord_id, buyer_user_id, 899.00, 'USD', 'credit_card', 'succeeded', 'txn_vnd_1008', NOW(), NOW())
        RETURNING id INTO pay_id;
        UPDATE orders SET payment_id = pay_id WHERE id = ord_id;
    END IF;

    RAISE NOTICE 'seed-orders-returns: vendor demo orders created for store luxe-atelier (user_id=%)', vendor_user_id;
END $$;
