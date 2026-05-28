-- =====================================================
-- Ensure we have at least some users (optional, skip if users exist)
-- =====================================================
-- If users table is empty, create some dummy users (adjust emails/passwords as needed)
INSERT INTO users (email, first_name, last_name, password_hash, role, is_active, created_at, updated_at)
SELECT
    'user' || i || '@example.com',
    'User' || i,
    'Test' || i,
    'hashed_password_placeholder', -- replace with real hash if needed
    'customer',
    true,
    NOW(),
    NOW()
FROM generate_series(1, 20) AS i
ON CONFLICT (email) DO NOTHING;

-- =====================================================
-- STORE FOLLOWERS (each store followed by 50-200 random users)
-- =====================================================
INSERT INTO store_followers (user_id, store_id, created_at)
SELECT
    u.id,
    s.id,
    NOW() - (random() * interval '180 days')
FROM
    (SELECT id FROM users LIMIT 200) u,
    (SELECT id FROM stores WHERE id BETWEEN 1 AND 18) s
WHERE
    random() < 0.3   -- each user follows ~30% of stores
ON CONFLICT (user_id, store_id) DO NOTHING;

-- =====================================================
-- STORE REVIEWS (each store gets 30-150 reviews)
-- =====================================================
INSERT INTO store_reviews (store_id, user_id, rating, comment, created_at, updated_at)
SELECT
    s.id,
    u.id,
    floor(random() * 5 + 1)::int,  -- rating 1-5
    CASE floor(random() * 5)::int
        WHEN 0 THEN 'Amazing store! Highly recommend.'
        WHEN 1 THEN 'Good products, fast shipping.'
        WHEN 2 THEN 'Decent experience, would shop again.'
        WHEN 3 THEN 'Average. Nothing special.'
        ELSE 'Could be better. Had some issues.'
        END,
    NOW() - (random() * interval '180 days'),
    NOW() - (random() * interval '180 days')
FROM
    (SELECT id FROM stores WHERE id BETWEEN 1 AND 18) s
        CROSS JOIN (SELECT id FROM users LIMIT 100) u
WHERE
    random() < 0.5   -- each user reviews ~50% of stores they "visited"
  -- Ensure a user can review a store only once
  AND NOT EXISTS (
    SELECT 1 FROM store_reviews sr WHERE sr.store_id = s.id AND sr.user_id = u.id
)
LIMIT 500;  -- safety, but we'll rely on ON CONFLICT later

-- If you want to ensure unique user+store combination, add a unique constraint:
-- ALTER TABLE store_reviews ADD CONSTRAINT unique_user_store_review UNIQUE (user_id, store_id);
-- But we'll simulate by using NOT EXISTS above.

-- Update store rating and review_count after seeding
UPDATE stores
SET
    review_count = (
        SELECT COUNT(*) FROM store_reviews WHERE store_reviews.store_id = stores.id
    ),
    rating = (
        SELECT COALESCE(AVG(rating), 0) FROM store_reviews WHERE store_reviews.store_id = stores.id
    )
WHERE id BETWEEN 1 AND 18;

-- Optionally, also update follower_count if needed
UPDATE stores
SET follower_count = (
    SELECT COUNT(*) FROM store_followers WHERE store_followers.store_id = stores.id
)
WHERE id BETWEEN 1 AND 18;