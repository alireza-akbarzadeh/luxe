-- Blog storefront demo posts (homepage + article detail).
-- Requires: blog_categories / blog_authors / blog_tags from migration 20260714120000.
-- Idempotent: ON CONFLICT (slug) DO UPDATE. Optional product links skip if no catalog.

-- ---------------------------------------------------------------------------
-- Featured: iPhone 16 Pro Max
-- ---------------------------------------------------------------------------
INSERT INTO blog_posts (
    slug, title, excerpt, hero_image_url, hero_image_alt, content_blocks,
    category_id, author_id, section_type, status,
    is_featured, is_editor_pick, is_trending,
    reading_time_minutes, view_count, helpful_votes,
    meta_title, meta_description, seo_score,
    published_at, content_updated_at, sort_order, created_at, updated_at
)
VALUES (
    'iphone-16-pro-max-review',
    'iPhone 16 Pro Max Review: The Best iPhone Yet?',
    'Apple''s flagship packs the A18 Pro chip, a dazzling camera system, and all-day battery — but does it justify the premium?',
    'https://images.unsplash.com/photo-1695048133142-1a20484d2569?w=1600&h=900&fit=crop',
    'iPhone 16 Pro Max front and back',
    $blocks$[
      {"type":"verdict","score":9.6,"label":"Quick Verdict","summary":"The iPhone 16 Pro Max is Apple’s most complete phone yet — exceptional display, camera, and battery, if you can live with the size and price."},
      {"type":"paragraph","text":"Apple’s latest Pro Max is a refinement play: brighter display, stronger silicon, and a camera stack that finally feels future-proof. We spent two weeks living with it as a daily driver."},
      {"type":"heading","level":2,"text":"Design & Build Quality"},
      {"type":"paragraph","text":"Titanium edges, Contour ceramic shield, and a camera plateau that looks intentional rather than bolted on. It is still a large phone — and that presence is part of the appeal."},
      {"type":"image","url":"https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=1200&h=675&fit=crop","alt":"iPhone side profile","caption":"Grade 5 titanium keeps the chassis feeling premium without excessive heft."},
      {"type":"heading","level":2,"text":"Display"},
      {"type":"paragraph","text":"The 6.9-inch LTPO OLED with ProMotion remains class-leading outdoors and for scrolling."},
      {"type":"list","style":"unordered","items":[
        {"text":"6.9-inch LTPO OLED with 120Hz ProMotion"},
        {"text":"Peak brightness suited for hard daylight"},
        {"text":"Always-On with proper dimming behavior"}
      ]},
      {"type":"heading","level":2,"text":"Performance"},
      {"type":"paragraph","text":"A18 Pro breezes through multitasking, games, and longer video exports. Thermals stay in check during sustained loads."},
      {"type":"table","headers":["Benchmark","iPhone 16 Pro Max","iPhone 15 Pro Max","Difference"],"rows":[
        ["Geekbench 6 Single","3500","2900","+21%"],
        ["Geekbench 6 Multi","8600","7200","+19%"],
        ["3DMark Wildlife Extreme","3800","3100","+23%"]
      ]},
      {"type":"heading","level":2,"text":"Camera"},
      {"type":"paragraph","text":"The 48MP main and ultra-wide duo capture better dynamic range, while the 5x telephoto covers most day-to-day zoom needs."},
      {"type":"pros_cons","pros":[
        "Stunning ProMotion display",
        "All-day battery with charge to spare",
        "Versatile camera system",
        "Excellent thermal performance"
      ],"cons":[
        "Very expensive starting price",
        "Large size may not suit everyone",
        "Incremental year-over-year design"
      ]},
      {"type":"faq","items":[
        {"question":"Is the iPhone 16 Pro Max worth upgrading to?","answer":"If you are on an iPhone 14 or older, yes — the camera, battery, and display gains are meaningful. From a 15 Pro Max, upgrades are subtler."},
        {"question":"How is the battery life on the iPhone 16 Pro Max?","answer":"Easily a full heavy day. Light users can stretch into a second calendar day without sweat."}
      ]},
      {"type":"cta","label":"Shop flagship smartphones","href":"/shop?category=smartphones"}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'smartphones' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'sarah-chen' LIMIT 1),
    'product_review',
    'published',
    TRUE, TRUE, TRUE,
    12, 87400, 2800,
    'iPhone 16 Pro Max Review: The Best iPhone Yet?',
    'Our full review of the iPhone 16 Pro Max — display, camera, performance, battery, and whether it is worth buying.',
    92,
    NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 day', 1, NOW(), NOW()
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    excerpt = EXCLUDED.excerpt,
    hero_image_url = EXCLUDED.hero_image_url,
    hero_image_alt = EXCLUDED.hero_image_alt,
    content_blocks = EXCLUDED.content_blocks,
    category_id = EXCLUDED.category_id,
    author_id = EXCLUDED.author_id,
    section_type = EXCLUDED.section_type,
    status = EXCLUDED.status,
    is_featured = EXCLUDED.is_featured,
    is_editor_pick = EXCLUDED.is_editor_pick,
    is_trending = EXCLUDED.is_trending,
    reading_time_minutes = EXCLUDED.reading_time_minutes,
    view_count = EXCLUDED.view_count,
    helpful_votes = EXCLUDED.helpful_votes,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    seo_score = EXCLUDED.seo_score,
    published_at = EXCLUDED.published_at,
    content_updated_at = EXCLUDED.content_updated_at,
    updated_at = NOW(),
    deleted_at = NULL;

-- ---------------------------------------------------------------------------
-- Detail focus: MacBook Air M4
-- ---------------------------------------------------------------------------
INSERT INTO blog_posts (
    slug, title, excerpt, hero_image_url, hero_image_alt, content_blocks,
    category_id, author_id, section_type, status,
    is_featured, is_editor_pick, is_trending,
    reading_time_minutes, view_count, helpful_votes,
    meta_title, meta_description, seo_score,
    published_at, content_updated_at, sort_order, created_at, updated_at
)
VALUES (
    'macbook-air-m4-review',
    'MacBook Air M4 Review: Feathery Power for Everyday Pros',
    'Apple’s thinnest laptop gets the M4 treatment — quieter fans (still none), brighter display options, and enough performance for creators who travel light.',
    'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=1600&h=900&fit=crop',
    'MacBook Air open on a desk',
    $blocks$[
      {"type":"verdict","score":9.4,"label":"Quick Verdict","summary":"The MacBook Air M4 is the default recommendation for most people who want a premium ultraportable: silent, fast, and all-day ready — with only GPU-heavy pros needing to look at the Pro lineup."},
      {"type":"paragraph","text":"The MacBook Air has long been Apple’s gateway laptop. With M4, it stops feeling like a compromise and starts feeling like the smarter buy for writers, students, developers, and light creators."},
      {"type":"heading","level":2,"text":"Design & Portability"},
      {"type":"paragraph","text":"Same featherweight chassis you already know, now with refined colors and a Liquid Retina panel that stays pleasant in cafes and airplanes. The MagSafe, twin Thunderbolt ports, and headphone jack cover modern essentials."},
      {"type":"image","url":"https://images.unsplash.com/photo-1611186871348-b1ce696e52c9?w=1200&h=675&fit=crop","alt":"MacBook Air closed lid","caption":"Under 1.3 kg in the 13-inch configuration — still the travel laptop to beat."},
      {"type":"heading","level":2,"text":"Display & Keyboard"},
      {"type":"list","style":"unordered","items":[
        {"text":"Bright Liquid Retina panel with P3 color"},
        {"text":"True Tone and wide viewing angles"},
        {"text":"Magic Keyboard with solid key travel"},
        {"text":"Large Force Touch trackpad that just works"}
      ]},
      {"type":"heading","level":2,"text":"Performance"},
      {"type":"paragraph","text":"M4 makes Xcode builds, browser tabs, and 4K timeline scrubs feel casual. Sustained workloads stay cool thanks to the efficient silicon — and that famous fanless design remains whisper-quiet."},
      {"type":"table","headers":["Workload","M4 Air","M3 Air","Notes"],"rows":[
        ["Xcode clean build","~18% faster","baseline","Noticeable on larger projects"],
        ["4K export (Final Cut)","~22% faster","baseline","Thermals stay under control"],
        ["Idle fan noise","Silent","Silent","Still no fans"}
      ]},
      {"type":"heading","level":2,"text":"Battery Life"},
      {"type":"paragraph","text":"Expect a full workday of browsing, docs, and video calls with reserve left. Streaming and coding all afternoon still leaves enough charge for the commute home."},
      {"type":"callout","tone":"tip","title":"Who should buy it","text":"Choose the Air if you value silence, weight, and value. Step up to a MacBook Pro if you edit multi-cam projects, run heavy local LLMs, or need more ports and peak brightness."},
      {"type":"pros_cons","pros":[
        "Silent fanless design",
        "Excellent everyday performance",
        "All-day battery life",
        "Gorgeously portable"
      ],"cons":[
        "Only two Thunderbolt ports",
        "No ProMotion display",
        "Base storage fills up quickly"
      ]},
      {"type":"faq","items":[
        {"question":"Is the MacBook Air M4 powerful enough for video editing?","answer":"For 4K cuts, color, and social exports — yes. Multi-cam or heavy effects work better on a MacBook Pro with active cooling."},
        {"question":"Should I get the 13-inch or 15-inch Air?","answer":"13-inch wins for backpacks and café tables. 15-inch is better if you want more screen real estate without jumping to a Pro."},
        {"question":"How does it compare to Windows ultraportables?","answer":"Battery life, silence, and macOS polish remain major advantages. Windows wins on ports, upgrades, and gaming."}
      ]},
      {"type":"cta","label":"Browse laptops","href":"/shop?category=laptops"}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'laptops' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'marcus-williams' LIMIT 1),
    'product_review',
    'published',
    FALSE, TRUE, TRUE,
    11, 45200, 1630,
    'MacBook Air M4 Review: Feathery Power for Everyday Pros',
    'In-depth MacBook Air M4 review covering design, performance, battery, pros and cons, and who should buy it.',
    90,
    NOW() - INTERVAL '2 days', NOW() - INTERVAL '12 hours', 2, NOW(), NOW()
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    excerpt = EXCLUDED.excerpt,
    hero_image_url = EXCLUDED.hero_image_url,
    hero_image_alt = EXCLUDED.hero_image_alt,
    content_blocks = EXCLUDED.content_blocks,
    category_id = EXCLUDED.category_id,
    author_id = EXCLUDED.author_id,
    section_type = EXCLUDED.section_type,
    status = EXCLUDED.status,
    is_featured = EXCLUDED.is_featured,
    is_editor_pick = EXCLUDED.is_editor_pick,
    is_trending = EXCLUDED.is_trending,
    reading_time_minutes = EXCLUDED.reading_time_minutes,
    view_count = EXCLUDED.view_count,
    helpful_votes = EXCLUDED.helpful_votes,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    seo_score = EXCLUDED.seo_score,
    published_at = EXCLUDED.published_at,
    content_updated_at = EXCLUDED.content_updated_at,
    updated_at = NOW(),
    deleted_at = NULL;

-- ---------------------------------------------------------------------------
-- Supporting posts (homepage sections)
-- ---------------------------------------------------------------------------
INSERT INTO blog_posts (
    slug, title, excerpt, hero_image_url, hero_image_alt, content_blocks,
    category_id, author_id, section_type, status,
    is_featured, is_editor_pick, is_trending,
    reading_time_minutes, view_count, helpful_votes,
    meta_title, meta_description, seo_score,
    published_at, content_updated_at, sort_order, created_at, updated_at
)
VALUES
(
    'best-ai-laptops-2026',
    'Best AI Laptops 2026: Copilot+ and Apple Silicon Compared',
    'We tested the year’s top AI-capable notebooks for on-device models, battery life, and real creative workflows.',
    'https://images.unsplash.com/photo-1496181133206-80ce9b88a853?w=1200&h=800&fit=crop',
    'Laptop on a wooden desk',
    $blocks$[
      {"type":"paragraph","text":"On-device AI finally matters for everyday laptops — but not every NPU claim translates to better UX."},
      {"type":"heading","level":2,"text":"How we tested"},
      {"type":"paragraph","text":"Battery rundowns, local model latency, and creator workloads across Windows Copilot+ PCs and Apple Silicon."},
      {"type":"cta","label":"Shop laptops","href":"/shop?category=laptops"}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'laptops' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'marcus-williams' LIMIT 1),
    'buying_guide', 'published', FALSE, TRUE, TRUE,
    9, 32100, 980,
    'Best AI Laptops 2026', 'Compare the best AI laptops for 2026.', 88,
    NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days', 3, NOW(), NOW()
),
(
    'sony-wh-1000xm6-vs-bose',
    'Sony WH-1000XM6 vs Bose QuietComfort Ultra',
    'Noise cancelling crowns clash again — which over-ears win for travel, calls, and music?',
    'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=1200&h=800&fit=crop',
    'Wireless headphones on a table',
    $blocks$[
      {"type":"paragraph","text":"Both pairs smother cabin drone; the differences show up in sound signature, call quality, and app polish."},
      {"type":"heading","level":2,"text":"Noise cancelling"},
      {"type":"paragraph","text":"Sony edges travel ANC; Bose feels more natural for voices in busy offices."},
      {"type":"pros_cons","pros":["Excellent ANC on both","Comfortable for long flights"],"cons":["Premium pricing","Cases still bulky"]}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'audio' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'elena-rodriguez' LIMIT 1),
    'comparison', 'published', FALSE, FALSE, TRUE,
    8, 28700, 740,
    'Sony WH-1000XM6 vs Bose QuietComfort Ultra', 'Head-to-head headphone comparison.', 85,
    NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days', 4, NOW(), NOW()
),
(
    'gaming-monitor-buying-guide',
    'Gaming Monitor Buying Guide: 144Hz to 4K OLED',
    'Refresh rate, panel type, HDMI 2.1 — cut through the marketing for your next display.',
    'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=1200&h=800&fit=crop',
    'Ultrawide gaming monitor',
    $blocks$[
      {"type":"paragraph","text":"Pick the panel for your platform first, then chase Hertz and HDR features."},
      {"type":"heading","level":2,"text":"Panel cheat sheet"},
      {"type":"list","style":"unordered","items":[
        {"text":"IPS for color accuracy"},
        {"text":"OLED for contrast and response"},
        {"text":"VA when you want deep blacks on a budget"}
      ]}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'gaming' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'marcus-williams' LIMIT 1),
    'buying_guide', 'published', FALSE, TRUE, FALSE,
    10, 19800, 520,
    'Gaming Monitor Buying Guide', 'How to choose a gaming monitor in 2026.', 84,
    NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days', 5, NOW(), NOW()
),
(
    'smart-home-starter-kit',
    'Smart Home Starter Kit Under $500',
    'Lights, locks, and a hub that does not fight you — a practical setup for renters and first-timers.',
    'https://images.unsplash.com/photo-1558002038-1055907dfabb?w=1200&h=800&fit=crop',
    'Smart home living room',
    $blocks$[
      {"type":"paragraph","text":"Start with bulbs and a matter-compatible hub — expand only after the basics feel invisible."},
      {"type":"heading","level":2,"text":"Shopping list"},
      {"type":"list","style":"ordered","items":[
        {"text":"Thread/Matter hub"},
        {"text":"Four smart bulbs"},
        {"text":"Door sensor or smart lock"},
        {"text":"One voice assistant speaker"}
      ]}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'smart-home' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'elena-rodriguez' LIMIT 1),
    'gift_guide', 'published', FALSE, FALSE, FALSE,
    7, 15400, 410,
    'Smart Home Starter Kit Under $500', 'Budget smart home starter guide.', 80,
    NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days', 6, NOW(), NOW()
),
(
    'how-to-calibrate-laptop-display',
    'How to Calibrate Your Laptop Display in 15 Minutes',
    'Quick steps for Windows and macOS so colors look consistent across your devices.',
    'https://images.unsplash.com/photo-1587825140708-dfaf72ae4b04?w=1200&h=800&fit=crop',
    'Color calibration on a laptop',
    $blocks$[
      {"type":"heading","level":2,"text":"macOS"},
      {"type":"paragraph","text":"Use System Settings → Displays → Color Profile, then validate with a known reference image."},
      {"type":"heading","level":2,"text":"Windows"},
      {"type":"paragraph","text":"Run Display Color Calibration and disable aggressive night-light modes while you adjust."}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'how-to-guides' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'sarah-chen' LIMIT 1),
    'tutorial', 'published', FALSE, FALSE, FALSE,
    6, 11200, 360,
    'How to Calibrate Your Laptop Display', 'Calibrate laptop display quickly on Mac and Windows.', 78,
    NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days', 7, NOW(), NOW()
),
(
    'chipmakers-race-2026',
    'Chipmakers Race into 2026: What Shoppers Should Watch',
    'NPU claims, process nodes, and why this year’s silicon changes what you should buy.',
    'https://images.unsplash.com/photo-1518770660439-4636190af475?w=1200&h=800&fit=crop',
    'Circuit board macro',
    $blocks$[
      {"type":"paragraph","text":"The marketing wars are louder than the real-world gaps — here is what actually moves battery life and AI features."},
      {"type":"heading","level":2,"text":"Buyer takeaways"},
      {"type":"list","style":"unordered","items":[
        {"text":"Prefer devices with mature driver support"},
        {"text":"Match NPU claims to apps you actually use"},
        {"text":"Battery software still matters more than peak TOP"}
      ]}
    ]$blocks$::jsonb,
    (SELECT id FROM blog_categories WHERE slug = 'industry-news' LIMIT 1),
    (SELECT id FROM blog_authors WHERE slug = 'sarah-chen' LIMIT 1),
    'industry_news', 'published', FALSE, FALSE, TRUE,
    5, 22100, 290,
    'Chipmakers Race into 2026', 'What silicon trends mean for shoppers in 2026.', 76,
    NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days', 8, NOW(), NOW()
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    excerpt = EXCLUDED.excerpt,
    hero_image_url = EXCLUDED.hero_image_url,
    hero_image_alt = EXCLUDED.hero_image_alt,
    content_blocks = EXCLUDED.content_blocks,
    category_id = EXCLUDED.category_id,
    author_id = EXCLUDED.author_id,
    section_type = EXCLUDED.section_type,
    status = EXCLUDED.status,
    is_featured = EXCLUDED.is_featured,
    is_editor_pick = EXCLUDED.is_editor_pick,
    is_trending = EXCLUDED.is_trending,
    reading_time_minutes = EXCLUDED.reading_time_minutes,
    view_count = EXCLUDED.view_count,
    helpful_votes = EXCLUDED.helpful_votes,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    seo_score = EXCLUDED.seo_score,
    published_at = EXCLUDED.published_at,
    content_updated_at = EXCLUDED.content_updated_at,
    updated_at = NOW(),
    deleted_at = NULL;

-- ---------------------------------------------------------------------------
-- Tags
-- ---------------------------------------------------------------------------
INSERT INTO blog_post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM blog_posts p
JOIN blog_tags t ON t.slug IN ('iphone', 'premium', '2026')
WHERE p.slug = 'iphone-16-pro-max-review'
ON CONFLICT DO NOTHING;

INSERT INTO blog_post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM blog_posts p
JOIN blog_tags t ON t.slug IN ('macbook', 'premium', '2026')
WHERE p.slug = 'macbook-air-m4-review'
ON CONFLICT DO NOTHING;

INSERT INTO blog_post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM blog_posts p
JOIN blog_tags t ON t.slug IN ('macbook', '2026')
WHERE p.slug = 'best-ai-laptops-2026'
ON CONFLICT DO NOTHING;

INSERT INTO blog_post_tags (post_id, tag_id)
SELECT p.id, t.id
FROM blog_posts p
JOIN blog_tags t ON t.slug IN ('wireless', 'premium')
WHERE p.slug = 'sony-wh-1000xm6-vs-bose'
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Optional product links (skip silently when catalog empty / no match)
-- ---------------------------------------------------------------------------
INSERT INTO blog_post_products (post_id, product_id, block_type, sort_order, created_at)
SELECT
    bp.id,
    p.id,
    'featured',
    0,
    NOW()
FROM blog_posts bp
JOIN products p ON p.deleted_at IS NULL AND p.status = 'active'
WHERE bp.slug = 'macbook-air-m4-review'
  AND (
      p.slug ILIKE '%macbook%'
      OR p.name ILIKE '%macbook%'
      OR p.name ILIKE '%laptop%'
  )
  AND NOT EXISTS (
      SELECT 1 FROM blog_post_products bpp
      WHERE bpp.post_id = bp.id AND bpp.product_id = p.id
  )
ORDER BY p.id
LIMIT 2;

INSERT INTO blog_post_products (post_id, product_id, block_type, sort_order, created_at)
SELECT
    bp.id,
    p.id,
    'featured',
    0,
    NOW()
FROM blog_posts bp
JOIN products p ON p.deleted_at IS NULL AND p.status = 'active'
WHERE bp.slug = 'iphone-16-pro-max-review'
  AND (
      p.slug ILIKE '%iphone%'
      OR p.name ILIKE '%iphone%'
      OR p.name ILIKE '%phone%'
  )
  AND NOT EXISTS (
      SELECT 1 FROM blog_post_products bpp
      WHERE bpp.post_id = bp.id AND bpp.product_id = p.id
  )
ORDER BY p.id
LIMIT 2;
