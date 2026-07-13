-- Real-world luxury fashion catalog for local/staging.
-- Run after seed-dev.sql (stores must exist). Idempotent: safe to re-run.

DO $$
DECLARE
  cat_women INT;
  cat_men INT;
  cat_accessories INT;
  cat_women_dresses INT;
  cat_women_tops INT;
  cat_women_knitwear INT;
  cat_women_outerwear INT;
  cat_women_shoes INT;
  cat_women_bags INT;
  cat_men_shirts INT;
  cat_men_trousers INT;
  cat_men_outerwear INT;
  cat_men_shoes INT;
  cat_watches INT;
  cat_jewelry INT;
  cat_sunglasses INT;
  cat_belts INT;
  brand_maison INT;
  brand_verona INT;
  brand_stellar INT;
  brand_atelier INT;
  brand_common INT;
  store_luxe INT;
  store_gold INT;
  store_urban INT;
  prod_id INT;
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'stores'
  ) THEN
    RAISE NOTICE 'seed-catalog: stores table missing — run make migrate-up first';
    RETURN;
  END IF;

  SELECT id INTO store_luxe FROM stores WHERE slug = 'luxe-atelier';
  SELECT id INTO store_gold FROM stores WHERE slug = 'gold-market';
  SELECT id INTO store_urban FROM stores WHERE slug = 'urban-essentials';

  IF store_luxe IS NULL THEN
    RAISE NOTICE 'seed-catalog: stores missing — run store seed first';
    RETURN;
  END IF;

  -- ── Brands ────────────────────────────────────────────────────────────────
  INSERT INTO brands (name, slug, description, logo_url, status, created_at, updated_at)
  VALUES
    ('Maison Éclat', 'maison-eclat', 'Parisian house known for silk tailoring and evening wear.', 'https://images.unsplash.com/photo-1560179707-f14e90ef3623?w=200', 'active', NOW(), NOW()),
    ('Verona Leather', 'verona-leather', 'Italian artisan leather goods since 1987.', 'https://images.unsplash.com/photo-1548036328-c9fa89d128fa?w=200', 'active', NOW(), NOW()),
    ('Stellar Time', 'stellar-time', 'Swiss-inspired timepieces with sapphire crystals.', 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=200', 'active', NOW(), NOW()),
    ('Atelier Nord', 'atelier-nord', 'Scandinavian minimal outerwear and knitwear.', 'https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=200', 'active', NOW(), NOW()),
    ('Common Thread', 'common-thread', 'Elevated everyday essentials in organic cotton.', 'https://images.unsplash.com/photo-1472851294608-062f824d29cc?w=200', 'active', NOW(), NOW())
  ON CONFLICT (slug) DO NOTHING;

  SELECT id INTO brand_maison FROM brands WHERE slug = 'maison-eclat';
  SELECT id INTO brand_verona FROM brands WHERE slug = 'verona-leather';
  SELECT id INTO brand_stellar FROM brands WHERE slug = 'stellar-time';
  SELECT id INTO brand_atelier FROM brands WHERE slug = 'atelier-nord';
  SELECT id INTO brand_common FROM brands WHERE slug = 'common-thread';

  -- ── Category helper: upsert by slug, fix level/path ───────────────────────
  CREATE OR REPLACE FUNCTION pg_temp.seed_category(
    p_name TEXT,
    p_slug TEXT,
    p_description TEXT,
    p_parent_id BIGINT
  ) RETURNS BIGINT AS $fn$
  DECLARE
    v_id BIGINT;
    v_level INT;
    v_path TEXT;
    v_parent_path TEXT;
    v_parent_level INT;
  BEGIN
    INSERT INTO categories (name, slug, description, parent_id, level, path, is_active, created_at, updated_at)
    VALUES (p_name, p_slug, p_description, p_parent_id, 0, '', TRUE, NOW(), NOW())
    ON CONFLICT (slug) DO NOTHING;

    SELECT id INTO v_id FROM categories WHERE slug = p_slug AND deleted_at IS NULL;

    IF p_parent_id IS NULL THEN
      v_level := 0;
      v_path := v_id::TEXT;
    ELSE
      SELECT level, path INTO v_parent_level, v_parent_path
      FROM categories WHERE id = p_parent_id;
      v_level := COALESCE(v_parent_level, 0) + 1;
      v_path := COALESCE(v_parent_path, p_parent_id::TEXT) || '.' || v_id::TEXT;
    END IF;

    UPDATE categories
    SET level = v_level, path = v_path, updated_at = NOW()
    WHERE id = v_id;

    RETURN v_id;
  END;
  $fn$ LANGUAGE plpgsql;

  cat_women := pg_temp.seed_category(
    'Women', 'women',
    'Contemporary womenswear — dresses, knitwear, shoes, and bags.',
    NULL
  );
  cat_men := pg_temp.seed_category(
    'Men', 'men',
    'Modern menswear — shirts, tailoring, outerwear, and footwear.',
    NULL
  );
  cat_accessories := pg_temp.seed_category(
    'Accessories', 'accessories',
    'Watches, jewelry, eyewear, and leather accessories.',
    NULL
  );

  cat_women_dresses := pg_temp.seed_category('Dresses', 'women-dresses', 'Midi, maxi, and occasion dresses.', cat_women);
  cat_women_tops := pg_temp.seed_category('Tops & Blouses', 'women-tops', 'Shirts, blouses, and lightweight layers.', cat_women);
  cat_women_knitwear := pg_temp.seed_category('Knitwear', 'women-knitwear', 'Cashmere, wool, and cotton knits.', cat_women);
  cat_women_outerwear := pg_temp.seed_category('Outerwear', 'women-outerwear', 'Coats, trenches, and jackets.', cat_women);
  cat_women_shoes := pg_temp.seed_category('Shoes', 'women-shoes', 'Heels, boots, flats, and sneakers.', cat_women);
  cat_women_bags := pg_temp.seed_category('Bags', 'women-bags', 'Totes, crossbody bags, and clutches.', cat_women);

  cat_men_shirts := pg_temp.seed_category('Shirts', 'men-shirts', 'Oxford, linen, and dress shirts.', cat_men);
  cat_men_trousers := pg_temp.seed_category('Trousers', 'men-trousers', 'Chinos, wool trousers, and denim.', cat_men);
  cat_men_outerwear := pg_temp.seed_category('Outerwear', 'men-outerwear', 'Blazers, coats, and overshirts.', cat_men);
  cat_men_shoes := pg_temp.seed_category('Shoes', 'men-shoes', 'Loafers, boots, and sneakers.', cat_men);

  cat_watches := pg_temp.seed_category('Watches', 'watches', 'Automatic, chronograph, and dress watches.', cat_accessories);
  cat_jewelry := pg_temp.seed_category('Jewelry', 'jewelry', 'Earrings, necklaces, and bracelets.', cat_accessories);
  cat_sunglasses := pg_temp.seed_category('Sunglasses', 'sunglasses', 'Polarized and UV-protective eyewear.', cat_accessories);
  cat_belts := pg_temp.seed_category('Belts & Scarves', 'belts-scarves', 'Leather belts and silk scarves.', cat_accessories);

  -- ── Products ────────────────────────────────────────────────────────────
  INSERT INTO products (
    name, slug, description, price, compare_at_price, cost, stock, sku, barcode,
    category_id, store_id, brand_id, status, rating, reviews_count, is_new,
    images, colors, sizes, tags, visibility, track_inventory, allow_backorder, weight,
    meta_title, meta_description, channels, published_at
  ) VALUES
  (
    'Arielle Silk Midi Dress',
    'arielle-silk-midi-dress',
    'Bias-cut silk midi dress with adjustable straps and concealed side zip. Fully lined. Dry clean only.',
    890.00, 1090.00, 310.00, 24, 'LUX-W-DRESS-001', '3700123456001',
    cat_women_dresses, store_luxe, brand_maison, 'active', 4.8, 42, TRUE,
    ARRAY['https://images.unsplash.com/photo-1595777457583-95e059d58199?w=900','https://images.unsplash.com/photo-1566174053879-31528523f8ae?w=900'],
    '["Black","Champagne","Emerald"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['dress','silk','evening','new-arrival'], 'public', TRUE, FALSE, 0.35,
    'Arielle Silk Midi Dress | Maison Éclat', 'Bias-cut silk midi dress in black, champagne, or emerald.',
    ARRAY['online_store'], NOW() - INTERVAL '14 days'
  ),
  (
    'Cloud Cashmere Crew Sweater',
    'cloud-cashmere-crew',
    'Grade-A Mongolian cashmere crew neck with ribbed cuffs. Lightweight enough for layering year-round.',
    425.00, 495.00, 145.00, 38, 'LUX-W-KNIT-001', '3700123456002',
    cat_women_knitwear, store_luxe, brand_atelier, 'active', 4.7, 31, TRUE,
    ARRAY['https://images.unsplash.com/photo-1576566588028-4147f3842f27?w=900'],
    '["Ivory","Camel","Charcoal"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['knitwear','cashmere','winter'], 'public', TRUE, FALSE, 0.28,
    'Cloud Cashmere Crew Sweater', 'Soft cashmere crew neck sweater in neutral tones.',
    ARRAY['online_store'], NOW() - INTERVAL '10 days'
  ),
  (
    'Milan Wool Double-Breasted Coat',
    'milan-wool-coat',
    'Full-length double-breasted coat in brushed Italian wool. Horn buttons, interior pocket, and half belt.',
    1290.00, 1490.00, 420.00, 15, 'LUX-W-COAT-001', '3700123456003',
    cat_women_outerwear, store_luxe, brand_maison, 'active', 4.9, 19, FALSE,
    ARRAY['https://images.unsplash.com/photo-1539533018447-63fcce267608?w=900','https://images.unsplash.com/photo-1487225417164-7164d20e98ae?w=900'],
    '["Camel","Navy","Black"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['coat','wool','outerwear'], 'public', TRUE, FALSE, 1.85,
    'Milan Wool Coat | Maison Éclat', 'Italian wool double-breasted coat for cold seasons.',
    ARRAY['online_store'], NOW() - INTERVAL '30 days'
  ),
  (
    'Verona Structured Leather Tote',
    'verona-structured-leather-tote',
    'Full-grain calfskin tote with magnetic closure, interior zip pocket, and detachable shoulder strap.',
    650.00, 750.00, 210.00, 20, 'LUX-W-BAG-001', '3700123456004',
    cat_women_bags, store_luxe, brand_verona, 'active', 4.6, 27, TRUE,
    ARRAY['https://images.unsplash.com/photo-1548036328-c9fa89d128fa?w=900','https://images.unsplash.com/photo-1584917865442-de89da76ac3a?w=900'],
    '["Tan","Black","Burgundy"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['bag','leather','handbag'], 'public', TRUE, FALSE, 0.92,
    'Verona Structured Leather Tote', 'Italian leather tote bag with shoulder strap.',
    ARRAY['online_store'], NOW() - INTERVAL '7 days'
  ),
  (
    'Arcadia Suede Ankle Boots',
    'arcadia-suede-ankle-boots',
    'Block-heel ankle boot in water-repellent suede with side zip and cushioned insole.',
    520.00, 590.00, 165.00, 22, 'LUX-W-SHOE-001', '3700123456005',
    cat_women_shoes, store_luxe, brand_verona, 'active', 4.5, 16, FALSE,
    ARRAY['https://images.unsplash.com/photo-1543163521-1bf539c55dd1?w=900'],
    '["Taupe","Black","Chocolate"]'::jsonb, '["36","37","38","39","40","41"]'::jsonb,
    ARRAY['boots','suede','footwear'], 'public', TRUE, FALSE, 0.78,
    'Arcadia Suede Ankle Boots', 'Suede block-heel ankle boots for everyday wear.',
    ARRAY['online_store'], NOW() - INTERVAL '21 days'
  ),
  (
    'Satin Pleated Midi Skirt',
    'satin-pleated-midi-skirt',
    'High-waist pleated midi skirt in lustrous satin with elasticated back waist.',
    295.00, 345.00, 88.00, 30, 'LUX-W-SKIRT-001', '3700123456006',
    cat_women_dresses, store_luxe, brand_maison, 'active', 4.4, 12, TRUE,
    ARRAY['https://images.unsplash.com/photo-1583498275914-65885544f1d0?w=900'],
    '["Blush","Navy","Black"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['skirt','satin','midi'], 'public', TRUE, FALSE, 0.22,
    'Satin Pleated Midi Skirt', 'Pleated satin midi skirt in blush, navy, or black.',
    ARRAY['online_store'], NOW() - INTERVAL '5 days'
  ),
  (
    'Classic Oxford Spread-Collar Shirt',
    'classic-oxford-shirt',
    'Long-staple cotton oxford shirt with spread collar and single cuff. Machine washable.',
    185.00, 215.00, 52.00, 45, 'LUX-M-SHIRT-001', '3700123456101',
    cat_men_shirts, store_urban, brand_common, 'active', 4.6, 58, FALSE,
    ARRAY['https://images.unsplash.com/photo-1602810318383-e386cc2a3ccf?w=900'],
    '["White","Light Blue","Pink"]'::jsonb, '["S","M","L","XL","XXL"]'::jsonb,
    ARRAY['shirt','oxford','essentials'], 'public', TRUE, FALSE, 0.24,
    'Classic Oxford Shirt | Common Thread', 'Cotton oxford shirt for office and weekend.',
    ARRAY['online_store','wholesale'], NOW() - INTERVAL '45 days'
  ),
  (
    'Navy Slim Wool Blazer',
    'navy-slim-wool-blazer',
    'Half-canvas construction with notch lapels, two-button closure, and functional sleeve buttons.',
    698.00, 798.00, 220.00, 18, 'LUX-M-BLAZER-001', '3700123456102',
    cat_men_outerwear, store_luxe, brand_maison, 'active', 4.7, 23, FALSE,
    ARRAY['https://images.unsplash.com/photo-1507679799987-c73779587ccf?w=900'],
    '["Navy","Charcoal"]'::jsonb, '["46","48","50","52","54"]'::jsonb,
    ARRAY['blazer','tailoring','wool'], 'public', TRUE, FALSE, 0.95,
    'Navy Slim Wool Blazer', 'Tailored wool blazer for business and events.',
    ARRAY['online_store'], NOW() - INTERVAL '20 days'
  ),
  (
    'Everyday Stretch Chino',
    'everyday-stretch-chino',
    'Garment-dyed stretch chino with hidden coin pocket and curved waistband.',
    145.00, 165.00, 38.00, 60, 'LUX-M-CHINO-001', '3700123456103',
    cat_men_trousers, store_urban, brand_common, 'active', 4.3, 74, FALSE,
    ARRAY['https://images.unsplash.com/photo-1473966968600-fa801b869a78?w=900'],
    '["Khaki","Navy","Olive","Stone"]'::jsonb, '["30","32","34","36","38"]'::jsonb,
    ARRAY['chino','trousers','casual'], 'public', TRUE, FALSE, 0.42,
    'Everyday Stretch Chino Trousers', 'Comfort stretch chinos in classic colors.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '60 days'
  ),
  (
    'Urban Chelsea Boot',
    'urban-chelsea-boot',
    'Pull-on Chelsea boot with Goodyear welt and natural rubber sole. Resolable construction.',
    395.00, 450.00, 118.00, 28, 'LUX-M-BOOT-001', '3700123456104',
    cat_men_shoes, store_urban, brand_verona, 'active', 4.5, 33, TRUE,
    ARRAY['https://images.unsplash.com/photo-1638247025967-b4e38f787b76?w=900'],
    '["Black","Cognac"]'::jsonb, '["40","41","42","43","44","45"]'::jsonb,
    ARRAY['boots','chelsea','leather'], 'public', TRUE, FALSE, 1.05,
    'Urban Chelsea Boot', 'Goodyear-welted Chelsea boots in black or cognac.',
    ARRAY['online_store'], NOW() - INTERVAL '12 days'
  ),
  (
    'Stellar Automatic 38mm',
    'stellar-automatic-38',
    'In-house automatic movement, domed sapphire crystal, and 50m water resistance. Exhibition caseback.',
    899.00, 1049.00, 280.00, 14, 'LUX-A-WATCH-001', '3700123456201',
    cat_watches, store_gold, brand_stellar, 'active', 4.8, 36, TRUE,
    ARRAY['https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=900','https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=900'],
    '["Silver","Rose Gold"]'::jsonb, '["38mm","40mm"]'::jsonb,
    ARRAY['watch','automatic','Swiss-style'], 'public', TRUE, FALSE, 0.38,
    'Stellar Automatic 38mm Watch', 'Automatic dress watch with sapphire crystal.',
    ARRAY['online_store'], NOW() - INTERVAL '18 days'
  ),
  (
    'Heritage Chronograph 42mm',
    'heritage-chronograph-42',
    'Sunburst dial chronograph with tachymeter bezel, sapphire glass, and quick-release leather strap.',
    1299.00, 1499.00, 360.00, 11, 'LUX-A-WATCH-002', '3700123456202',
    cat_watches, store_gold, brand_stellar, 'active', 4.7, 28, FALSE,
    ARRAY['https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=900'],
    '["Silver","Gold","Blue"]'::jsonb, '["42mm"]'::jsonb,
    ARRAY['watch','chronograph','heritage'], 'public', TRUE, FALSE, 0.44,
    'Heritage Chronograph 42mm', 'Swiss-style chronograph with sunburst dial.',
    ARRAY['online_store'], NOW() - INTERVAL '40 days'
  ),
  (
    'Solstice Gold Vermeil Hoops',
    'solstice-gold-vermeil-hoops',
    'Medium hoop earrings in 18k gold vermeil over sterling silver. Hypoallergenic posts.',
    320.00, 360.00, 95.00, 40, 'LUX-A-JEWEL-001', '3700123456203',
    cat_jewelry, store_gold, NULL, 'active', 4.6, 21, TRUE,
    ARRAY['https://images.unsplash.com/photo-1515562141207-29a036fb126a?w=900'],
    '["Gold"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['jewelry','earrings','gold'], 'public', TRUE, FALSE, 0.04,
    'Solstice Gold Vermeil Hoops', 'Gold vermeil hoop earrings.',
    ARRAY['online_store'], NOW() - INTERVAL '8 days'
  ),
  (
    'Horizon Polarized Aviators',
    'horizon-polarized-aviators',
    'Lightweight titanium aviators with polarized CR-39 lenses and adjustable nose pads.',
    245.00, 285.00, 68.00, 35, 'LUX-A-SUN-001', '3700123456204',
    cat_sunglasses, store_luxe, NULL, 'active', 4.4, 17, FALSE,
    ARRAY['https://images.unsplash.com/photo-1572635196233-4b24fbf0d9d9?w=900'],
    '["Gold/Green","Silver/Grey","Black/Blue"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['sunglasses','aviator','polarized'], 'public', TRUE, FALSE, 0.06,
    'Horizon Polarized Aviators', 'Titanium aviator sunglasses with polarized lenses.',
    ARRAY['online_store'], NOW() - INTERVAL '25 days'
  ),
  (
    'Milano Calfskin Belt',
    'milano-calfskin-belt',
    '1.25-inch calfskin belt with brushed nickel buckle. Sized from 32 to 42.',
    165.00, 185.00, 42.00, 50, 'LUX-A-BELT-001', '3700123456205',
    cat_belts, store_luxe, brand_verona, 'active', 4.5, 14, FALSE,
    ARRAY['https://images.unsplash.com/photo-1624222247344-550fb60583fd?w=900'],
    '["Black","Brown","Cognac"]'::jsonb, '["32","34","36","38","40","42"]'::jsonb,
    ARRAY['belt','leather','accessory'], 'public', TRUE, FALSE, 0.18,
    'Milano Calfskin Belt', 'Italian calfskin belt with nickel buckle.',
    ARRAY['online_store'], NOW() - INTERVAL '35 days'
  ),
  (
    'Essential Organic Cotton Tee',
    'essential-organic-cotton-tee',
    'GOTS-certified organic cotton jersey tee with reinforced neckline. Pre-shrunk.',
    48.00, 58.00, 12.00, 120, 'LUX-M-TEE-001', '3700123456105',
    cat_men_shirts, store_urban, brand_common, 'active', 4.2, 112, FALSE,
    ARRAY['https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=900'],
    '["White","Black","Heather Grey","Navy"]'::jsonb, '["S","M","L","XL","XXL"]'::jsonb,
    ARRAY['tee','basics','cotton'], 'public', TRUE, TRUE, 0.18,
    'Essential Organic Cotton Tee', 'Everyday organic cotton t-shirt.',
    ARRAY['online_store','pos','wholesale'], NOW() - INTERVAL '90 days'
  ),
  (
    'Mini Quilted Crossbody',
    'mini-quilted-crossbody',
    'Diamond-quilted crossbody with chain strap and microfiber lining. Fits phone and card holder.',
    425.00, 475.00, 130.00, 26, 'LUX-W-BAG-002', '3700123456007',
    cat_women_bags, store_luxe, brand_verona, 'active', 4.7, 20, TRUE,
    ARRAY['https://images.unsplash.com/photo-1564422170191-4bd349468e38?w=900'],
    '["Black","Cream","Red"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['bag','crossbody','quilted'], 'public', TRUE, FALSE, 0.35,
    'Mini Quilted Crossbody Bag', 'Quilted leather crossbody with chain strap.',
    ARRAY['online_store'], NOW() - INTERVAL '6 days'
  ),
  (
    'Merino Roll Neck Sweater',
    'merino-roll-neck-sweater',
    'Fine merino wool roll neck with fully fashioned seams. Ideal base layer or standalone piece.',
    198.00, 228.00, 55.00, 42, 'LUX-W-KNIT-002', '3700123456008',
    cat_women_knitwear, store_luxe, brand_atelier, 'active', 4.5, 15, FALSE,
    ARRAY['https://images.unsplash.com/photo-1434389677669-e08b4cac3105?w=900'],
    '["Black","Grey Melange","Forest"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['knitwear','merino','layering'], 'public', TRUE, FALSE, 0.26,
    'Merino Roll Neck Sweater', 'Fine merino roll neck sweater.',
    ARRAY['online_store'], NOW() - INTERVAL '16 days'
  ),
  (
    'Linen Camp Collar Shirt',
    'linen-camp-collar-shirt',
    'Relaxed camp-collar shirt in garment-washed European linen. Corozo buttons.',
    165.00, 185.00, 44.00, 34, 'LUX-M-SHIRT-002', '3700123456106',
    cat_men_shirts, store_luxe, brand_atelier, 'active', 4.4, 19, TRUE,
    ARRAY['https://images.unsplash.com/photo-1596755094514-f87e34085b23?w=900'],
    '["White","Sand","Sky Blue"]'::jsonb, '["S","M","L","XL"]'::jsonb,
    ARRAY['shirt','linen','summer'], 'public', TRUE, FALSE, 0.21,
    'Linen Camp Collar Shirt', 'Relaxed linen shirt for warm weather.',
    ARRAY['online_store'], NOW() - INTERVAL '9 days'
  ),
  -- ── Extra catalog volume (pagination, category filters, empty leaves) ──
  (
    'Poplin Wrap Blouse',
    'poplin-wrap-blouse',
    'Soft cotton-poplin wrap blouse with self-tie waist and gently puffed sleeves.',
    210.00, 245.00, 62.00, 36, 'LUX-W-TOP-001', '3700123456009',
    cat_women_tops, store_luxe, brand_maison, 'active', 4.5, 18, TRUE,
    ARRAY['https://images.unsplash.com/photo-1564257631407-4deb1f99d992?w=900'],
    '["Ivory","Black","Soft Rose"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['blouse','tops','office'], 'public', TRUE, FALSE, 0.19,
    'Poplin Wrap Blouse', 'Cotton wrap blouse for work and weekend.',
    ARRAY['online_store'], NOW() - INTERVAL '4 days'
  ),
  (
    'Silk Cap-Sleeve Shell',
    'silk-cap-sleeve-shell',
    'Bias-cut silk shell with clean boat neck and invisible zip. Perfect under blazers.',
    275.00, 310.00, 78.00, 28, 'LUX-W-TOP-002', '3700123456010',
    cat_women_tops, store_luxe, brand_maison, 'active', 4.6, 11, FALSE,
    ARRAY['https://images.unsplash.com/photo-1594633312681-425c7b97ccd1?w=900'],
    '["Black","Champagne","Navy"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['silk','shell','tops'], 'public', TRUE, FALSE, 0.12,
    'Silk Cap-Sleeve Shell', 'Minimal silk shell in evening neutrals.',
    ARRAY['online_store'], NOW() - INTERVAL '11 days'
  ),
  (
    'Relaxed Linen Button-Up',
    'relaxed-linen-button-up',
    'Oversized linen button-up with mother-of-pearl buttons and curved hem.',
    185.00, 210.00, 48.00, 40, 'LUX-W-TOP-003', '3700123456011',
    cat_women_tops, store_urban, brand_atelier, 'active', 4.3, 22, TRUE,
    ARRAY['https://images.unsplash.com/photo-1485968579580-b6d095142e6e?w=900'],
    '["White","Sage","Sand"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['linen','shirt','summer','tops'], 'public', TRUE, FALSE, 0.20,
    'Relaxed Linen Button-Up', 'Oversized linen shirt in soft summer tones.',
    ARRAY['online_store'], NOW() - INTERVAL '3 days'
  ),
  (
    'Ribbed Mock-Neck Tank',
    'ribbed-mock-neck-tank',
    'Fine-rib stretch tank with mock neck. Layer under knits or wear alone.',
    95.00, 110.00, 24.00, 55, 'LUX-W-TOP-004', '3700123456012',
    cat_women_tops, store_urban, brand_common, 'active', 4.2, 34, FALSE,
    ARRAY['https://images.unsplash.com/photo-1503342217505-b0a15ec3261c?w=900'],
    '["Black","White","Espresso"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['tank','basics','tops'], 'public', TRUE, TRUE, 0.10,
    'Ribbed Mock-Neck Tank', 'Everyday ribbed tank in core colors.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '28 days'
  ),
  (
    'Velvet Column Evening Dress',
    'velvet-column-evening-dress',
    'Floor-skimming velvet column with square neckline and open back.',
    1180.00, 1380.00, 390.00, 12, 'LUX-W-DRESS-002', '3700123456013',
    cat_women_dresses, store_luxe, brand_maison, 'active', 4.9, 14, TRUE,
    ARRAY['https://images.unsplash.com/photo-1515372039744-b8f02a3ae446?w=900'],
    '["Burgundy","Black","Forest"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['dress','velvet','evening','new-arrival'], 'public', TRUE, FALSE, 0.55,
    'Velvet Column Evening Dress', 'Statement velvet evening dress.',
    ARRAY['online_store'], NOW() - INTERVAL '2 days'
  ),
  (
    'Daylight Cotton Shirt Dress',
    'daylight-cotton-shirt-dress',
    'Crisp cotton shirt dress with belted waist and side pockets.',
    340.00, 390.00, 95.00, 32, 'LUX-W-DRESS-003', '3700123456014',
    cat_women_dresses, store_urban, brand_common, 'active', 4.4, 26, FALSE,
    ARRAY['https://images.unsplash.com/photo-1496747611176-843222e1e57c?w=900'],
    '["White","Stripe Navy","Khaki"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['dress','shirt-dress','daywear'], 'public', TRUE, FALSE, 0.30,
    'Daylight Cotton Shirt Dress', 'Easy cotton shirt dress for daytime.',
    ARRAY['online_store'], NOW() - INTERVAL '19 days'
  ),
  (
    'Alpaca Oversized Cardigan',
    'alpaca-oversized-cardigan',
    'Brushed alpaca blend cardigan with patch pockets and drop shoulders.',
    485.00, 545.00, 155.00, 22, 'LUX-W-KNIT-003', '3700123456015',
    cat_women_knitwear, store_luxe, brand_atelier, 'active', 4.7, 17, TRUE,
    ARRAY['https://images.unsplash.com/photo-1434389677669-e08b4cac3105?w=900'],
    '["Oatmeal","Smoke","Wine"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['cardigan','alpaca','knitwear'], 'public', TRUE, FALSE, 0.48,
    'Alpaca Oversized Cardigan', 'Soft oversized cardigan for cool evenings.',
    ARRAY['online_store'], NOW() - INTERVAL '13 days'
  ),
  (
    'Cropped Cashmere Polo',
    'cropped-cashmere-polo',
    'Cropped cashmere polo with mother-of-pearl buttons and rib hem.',
    395.00, 445.00, 130.00, 25, 'LUX-W-KNIT-004', '3700123456016',
    cat_women_knitwear, store_luxe, brand_atelier, 'active', 4.6, 13, FALSE,
    ARRAY['https://images.unsplash.com/photo-1620799140408-edc6dcb6d633?w=900'],
    '["Ivory","Dusty Pink","Black"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['polo','cashmere','knitwear'], 'public', TRUE, FALSE, 0.22,
    'Cropped Cashmere Polo', 'Cropped cashmere polo sweater.',
    ARRAY['online_store'], NOW() - INTERVAL '17 days'
  ),
  (
    'City Rain Trench',
    'city-rain-trench',
    'Water-repellent cotton trench with removable lining and storm flap.',
    890.00, 990.00, 280.00, 16, 'LUX-W-COAT-002', '3700123456017',
    cat_women_outerwear, store_luxe, brand_atelier, 'active', 4.8, 21, FALSE,
    ARRAY['https://images.unsplash.com/photo-1548126032-077a2e8e9e3b?w=900'],
    '["Stone","Black","Olive"]'::jsonb, '["XS","S","M","L","XL"]'::jsonb,
    ARRAY['trench','outerwear','rain'], 'public', TRUE, FALSE, 1.10,
    'City Rain Trench', 'Classic trench for wet city days.',
    ARRAY['online_store'], NOW() - INTERVAL '33 days'
  ),
  (
    'Soft Leather Biker Jacket',
    'soft-leather-biker-jacket',
    'Lambskin biker with asymmetric zip, quilted shoulders, and silver hardware.',
    1450.00, 1650.00, 480.00, 10, 'LUX-W-COAT-003', '3700123456018',
    cat_women_outerwear, store_luxe, brand_verona, 'active', 4.9, 9, TRUE,
    ARRAY['https://images.unsplash.com/photo-1551028719-00167b16eac5?w=900'],
    '["Black","Cognac"]'::jsonb, '["XS","S","M","L"]'::jsonb,
    ARRAY['leather','jacket','outerwear','new-arrival'], 'public', TRUE, FALSE, 1.20,
    'Soft Leather Biker Jacket', 'Lambskin biker jacket with silver hardware.',
    ARRAY['online_store'], NOW() - INTERVAL '1 day'
  ),
  (
    'Sculptural Leather Pump',
    'sculptural-leather-pump',
    'Pointed-toe pump with 75mm sculptural heel and kid leather upper.',
    480.00, 540.00, 150.00, 27, 'LUX-W-SHOE-002', '3700123456019',
    cat_women_shoes, store_luxe, brand_verona, 'active', 4.5, 24, FALSE,
    ARRAY['https://images.unsplash.com/photo-1543163521-1bf539c55dd1?w=900'],
    '["Black","Nude","Red"]'::jsonb, '["36","37","38","39","40"]'::jsonb,
    ARRAY['heels','pump','footwear'], 'public', TRUE, FALSE, 0.55,
    'Sculptural Leather Pump', 'Elegant pointed pump for evening and office.',
    ARRAY['online_store'], NOW() - INTERVAL '22 days'
  ),
  (
    'Cloud Knit Sneaker',
    'cloud-knit-sneaker',
    'Lightweight knit upper sneaker with cushioned foam sole and reflective lace tips.',
    220.00, 250.00, 65.00, 48, 'LUX-W-SHOE-003', '3700123456020',
    cat_women_shoes, store_urban, brand_common, 'active', 4.3, 41, TRUE,
    ARRAY['https://images.unsplash.com/photo-1549298916-b41d501d3772?w=900'],
    '["White","Black","Grey"]'::jsonb, '["36","37","38","39","40","41"]'::jsonb,
    ARRAY['sneakers','knit','footwear'], 'public', TRUE, FALSE, 0.45,
    'Cloud Knit Sneaker', 'Everyday knit sneakers with cloud cushioning.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '8 days'
  ),
  (
    'Weekend Canvas Tote',
    'weekend-canvas-tote',
    'Heavyweight canvas tote with leather handles and interior laptop sleeve.',
    185.00, 210.00, 48.00, 44, 'LUX-W-BAG-003', '3700123456021',
    cat_women_bags, store_urban, brand_common, 'active', 4.2, 29, FALSE,
    ARRAY['https://images.unsplash.com/photo-1590874103328-eac38a683ce7?w=900'],
    '["Natural","Black","Navy"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['tote','canvas','bag'], 'public', TRUE, FALSE, 0.70,
    'Weekend Canvas Tote', 'Roomy canvas tote for work and travel.',
    ARRAY['online_store'], NOW() - INTERVAL '27 days'
  ),
  (
    'Evening Satin Clutch',
    'evening-satin-clutch',
    'Fold-over satin clutch with magnetic snap and detachable chain.',
    265.00, 295.00, 72.00, 30, 'LUX-W-BAG-004', '3700123456022',
    cat_women_bags, store_luxe, brand_maison, 'active', 4.6, 15, TRUE,
    ARRAY['https://images.unsplash.com/photo-1566150905458-1bf1fc113f0d?w=900'],
    '["Black","Gold","Emerald"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['clutch','evening','bag'], 'public', TRUE, FALSE, 0.25,
    'Evening Satin Clutch', 'Compact satin clutch for nights out.',
    ARRAY['online_store'], NOW() - INTERVAL '6 days'
  ),
  (
    'French Blue Dress Shirt',
    'french-blue-dress-shirt',
    'Twisted yarn cotton dress shirt with French cuffs and fused collar.',
    225.00, 255.00, 58.00, 38, 'LUX-M-SHIRT-003', '3700123456107',
    cat_men_shirts, store_luxe, brand_maison, 'active', 4.7, 31, FALSE,
    ARRAY['https://images.unsplash.com/photo-1598033129183-c4f50c736f10?w=900'],
    '["French Blue","White","Lilac"]'::jsonb, '["S","M","L","XL"]'::jsonb,
    ARRAY['shirt','dress','formal'], 'public', TRUE, FALSE, 0.26,
    'French Blue Dress Shirt', 'Formal cotton dress shirt with French cuffs.',
    ARRAY['online_store'], NOW() - INTERVAL '24 days'
  ),
  (
    'Heavyweight Jersey Hoodie',
    'heavyweight-jersey-hoodie',
    '450gsm loopback hoodie with kangaroo pocket and tonal drawcord.',
    165.00, 185.00, 42.00, 52, 'LUX-M-SHIRT-004', '3700123456108',
    cat_men_shirts, store_urban, brand_common, 'active', 4.4, 67, TRUE,
    ARRAY['https://images.unsplash.com/photo-1556821840-3a63f95609a7?w=900'],
    '["Black","Grey","Forest"]'::jsonb, '["S","M","L","XL","XXL"]'::jsonb,
    ARRAY['hoodie','jersey','casual'], 'public', TRUE, TRUE, 0.65,
    'Heavyweight Jersey Hoodie', 'Premium heavyweight hoodie for everyday.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '5 days'
  ),
  (
    'Pleated Wool Trouser',
    'pleated-wool-trouser',
    'Single-pleat wool trouser with side adjusters and unfinished hem.',
    320.00, 360.00, 95.00, 24, 'LUX-M-TROU-002', '3700123456109',
    cat_men_trousers, store_luxe, brand_maison, 'active', 4.6, 18, FALSE,
    ARRAY['https://images.unsplash.com/photo-1473966968600-fa801b869a78?w=900'],
    '["Charcoal","Navy","Camel"]'::jsonb, '["30","32","34","36","38"]'::jsonb,
    ARRAY['trousers','wool','tailoring'], 'public', TRUE, FALSE, 0.50,
    'Pleated Wool Trouser', 'Tailored wool trousers with unfinished hem.',
    ARRAY['online_store'], NOW() - INTERVAL '29 days'
  ),
  (
    'Selvedge Straight Denim',
    'selvedge-straight-denim',
    '14oz Japanese selvedge denim with button fly and clean finish.',
    245.00, 275.00, 72.00, 40, 'LUX-M-TROU-003', '3700123456110',
    cat_men_trousers, store_urban, brand_common, 'active', 4.5, 52, TRUE,
    ARRAY['https://images.unsplash.com/photo-1542272454315-7f6b4f5d0f0b?w=900'],
    '["Indigo","Black"]'::jsonb, '["30","32","34","36"]'::jsonb,
    ARRAY['denim','jeans','selvedge'], 'public', TRUE, FALSE, 0.70,
    'Selvedge Straight Denim', 'Straight-leg Japanese selvedge jeans.',
    ARRAY['online_store'], NOW() - INTERVAL '10 days'
  ),
  (
    'Field Cotton Overshirt',
    'field-cotton-overshirt',
    'Garment-dyed cotton overshirt with twin chest pockets and horn buttons.',
    275.00, 310.00, 78.00, 29, 'LUX-M-OUT-002', '3700123456111',
    cat_men_outerwear, store_urban, brand_atelier, 'active', 4.4, 20, FALSE,
    ARRAY['https://images.unsplash.com/photo-1591047139829-d91aecb6caea?w=900'],
    '["Olive","Sand","Navy"]'::jsonb, '["S","M","L","XL"]'::jsonb,
    ARRAY['overshirt','outerwear','layering'], 'public', TRUE, FALSE, 0.55,
    'Field Cotton Overshirt', 'Utility overshirt for transitional weather.',
    ARRAY['online_store'], NOW() - INTERVAL '15 days'
  ),
  (
    'Quilted Packable Vest',
    'quilted-packable-vest',
    'Ultralight quilted vest that packs into its own pocket. Wind-resistant shell.',
    195.00, 225.00, 55.00, 35, 'LUX-M-OUT-003', '3700123456112',
    cat_men_outerwear, store_urban, brand_atelier, 'active', 4.3, 27, TRUE,
    ARRAY['https://images.unsplash.com/photo-1551028719-00167b16eac5?w=900'],
    '["Black","Navy","Burnt Orange"]'::jsonb, '["S","M","L","XL"]'::jsonb,
    ARRAY['vest','quilted','outerwear'], 'public', TRUE, FALSE, 0.35,
    'Quilted Packable Vest', 'Lightweight packable quilted vest.',
    ARRAY['online_store'], NOW() - INTERVAL '7 days'
  ),
  (
    'Hand-Finished Penny Loafer',
    'hand-finished-penny-loafer',
    'Burnished calf penny loafer with leather sole and stacked heel.',
    420.00, 475.00, 125.00, 26, 'LUX-M-SHOE-002', '3700123456113',
    cat_men_shoes, store_luxe, brand_verona, 'active', 4.7, 22, FALSE,
    ARRAY['https://images.unsplash.com/photo-1614252369475-531eba835eb1?w=900'],
    '["Black","Cognac","Burgundy"]'::jsonb, '["40","41","42","43","44","45"]'::jsonb,
    ARRAY['loafer','leather','footwear'], 'public', TRUE, FALSE, 0.85,
    'Hand-Finished Penny Loafer', 'Classic penny loafers in burnished calf.',
    ARRAY['online_store'], NOW() - INTERVAL '31 days'
  ),
  (
    'Court Leather Sneaker',
    'court-leather-sneaker',
    'Low-profile leather court sneaker with gum sole and perforated toe.',
    285.00, 320.00, 82.00, 42, 'LUX-M-SHOE-003', '3700123456114',
    cat_men_shoes, store_urban, brand_common, 'active', 4.4, 38, TRUE,
    ARRAY['https://images.unsplash.com/photo-1525966222134-fcfa99b8ae77?w=900'],
    '["White","Navy","Black"]'::jsonb, '["40","41","42","43","44","45"]'::jsonb,
    ARRAY['sneakers','leather','footwear'], 'public', TRUE, FALSE, 0.60,
    'Court Leather Sneaker', 'Minimal leather sneakers with gum sole.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '9 days'
  ),
  (
    'Dive Automatic 42mm',
    'dive-automatic-42',
    '200m dive watch with ceramic bezel, lume markers, and screw-down crown.',
    1099.00, 1249.00, 320.00, 13, 'LUX-A-WATCH-003', '3700123456206',
    cat_watches, store_gold, brand_stellar, 'active', 4.8, 25, TRUE,
    ARRAY['https://images.unsplash.com/photo-1524592094714-0f0654e20314?w=900'],
    '["Black","Blue"]'::jsonb, '["42mm"]'::jsonb,
    ARRAY['watch','dive','automatic'], 'public', TRUE, FALSE, 0.52,
    'Dive Automatic 42mm', 'Ceramic-bezel dive watch with 200m rating.',
    ARRAY['online_store'], NOW() - INTERVAL '14 days'
  ),
  (
    'Minimal Field Watch 36mm',
    'minimal-field-watch-36',
    'Slim field watch with matte dial, sapphire crystal, and NATO strap options.',
    549.00, 629.00, 160.00, 20, 'LUX-A-WATCH-004', '3700123456207',
    cat_watches, store_gold, brand_stellar, 'active', 4.5, 19, FALSE,
    ARRAY['https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=900'],
    '["Olive","Sand","Black"]'::jsonb, '["36mm","38mm"]'::jsonb,
    ARRAY['watch','field','minimal'], 'public', TRUE, FALSE, 0.32,
    'Minimal Field Watch 36mm', 'Everyday field watch with sapphire crystal.',
    ARRAY['online_store'], NOW() - INTERVAL '26 days'
  ),
  (
    'Layered Pearl Pendant',
    'layered-pearl-pendant',
    'Freshwater pearl on adjustable gold-filled chain with satellite bead.',
    185.00, 210.00, 48.00, 45, 'LUX-A-JEWEL-002', '3700123456208',
    cat_jewelry, store_gold, NULL, 'active', 4.4, 28, TRUE,
    ARRAY['https://images.unsplash.com/photo-1599643478518-a784e5dc4c8f?w=900'],
    '["Gold","Silver"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['jewelry','necklace','pearl'], 'public', TRUE, FALSE, 0.03,
    'Layered Pearl Pendant', 'Delicate pearl pendant necklace.',
    ARRAY['online_store'], NOW() - INTERVAL '12 days'
  ),
  (
    'Stackable Signet Ring',
    'stackable-signet-ring',
    'Brushed sterling signet designed for stacking. Engravable face.',
    145.00, 165.00, 38.00, 50, 'LUX-A-JEWEL-003', '3700123456209',
    cat_jewelry, store_gold, NULL, 'active', 4.3, 16, FALSE,
    ARRAY['https://images.unsplash.com/photo-1605100804763-247f67b3557e?w=900'],
    '["Silver","Gold"]'::jsonb, '["5","6","7","8","9"]'::jsonb,
    ARRAY['jewelry','ring','signet'], 'public', TRUE, FALSE, 0.02,
    'Stackable Signet Ring', 'Sterling signet ring for everyday stacking.',
    ARRAY['online_store'], NOW() - INTERVAL '20 days'
  ),
  (
    'Acetate Square Sunglasses',
    'acetate-square-sunglasses',
    'Hand-polished acetate frames with UV400 lenses and metal hinge core.',
    195.00, 225.00, 52.00, 38, 'LUX-A-SUN-002', '3700123456210',
    cat_sunglasses, store_luxe, NULL, 'active', 4.5, 23, TRUE,
    ARRAY['https://images.unsplash.com/photo-1511499767150-a48a237f0083?w=900'],
    '["Tortoise","Black","Clear"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['sunglasses','acetate','square'], 'public', TRUE, FALSE, 0.05,
    'Acetate Square Sunglasses', 'Square acetate sunglasses with UV400 lenses.',
    ARRAY['online_store'], NOW() - INTERVAL '11 days'
  ),
  (
    'Silk Square Scarf 90',
    'silk-square-scarf-90',
    '90cm silk twill scarf with hand-rolled edges and archival print.',
    275.00, 310.00, 70.00, 33, 'LUX-A-BELT-002', '3700123456211',
    cat_belts, store_luxe, brand_maison, 'active', 4.7, 12, TRUE,
    ARRAY['https://images.unsplash.com/photo-1601924999987-bde7dbbb1691?w=900'],
    '["Multicolor","Navy","Ivory"]'::jsonb, '["One Size"]'::jsonb,
    ARRAY['scarf','silk','accessory'], 'public', TRUE, FALSE, 0.08,
    'Silk Square Scarf 90', 'Hand-rolled silk twill scarf.',
    ARRAY['online_store'], NOW() - INTERVAL '16 days'
  ),
  (
    'Reversible Leather Belt',
    'reversible-leather-belt',
    'Reversible black/brown calf belt with rotating buckle. One belt, two looks.',
    145.00, 165.00, 36.00, 55, 'LUX-A-BELT-003', '3700123456212',
    cat_belts, store_urban, brand_verona, 'active', 4.4, 30, FALSE,
    ARRAY['https://images.unsplash.com/photo-1624222247344-550fb60583fd?w=900'],
    '["Black/Brown"]'::jsonb, '["32","34","36","38","40","42"]'::jsonb,
    ARRAY['belt','reversible','leather'], 'public', TRUE, FALSE, 0.20,
    'Reversible Leather Belt', 'Black and brown reversible calfskin belt.',
    ARRAY['online_store','pos'], NOW() - INTERVAL '38 days'
  )
  ON CONFLICT (slug) DO NOTHING;

  -- ── Product attributes (color / size pickers) for all active storefront SKUs ─
  FOR prod_id IN
    SELECT id FROM products
    WHERE store_id IN (store_luxe, store_gold, store_urban)
      AND status = 'active'
      AND visibility = 'public'
  LOOP
    IF NOT EXISTS (SELECT 1 FROM product_attributes WHERE product_id = prod_id AND name = 'color') THEN
      INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
      SELECT prod_id, 'color', ARRAY(SELECT jsonb_array_elements_text(colors)), NOW(), NOW()
      FROM products WHERE id = prod_id AND colors IS NOT NULL AND colors::text <> '[]';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM product_attributes WHERE product_id = prod_id AND name = 'size') THEN
      INSERT INTO product_attributes (product_id, name, values, created_at, updated_at)
      SELECT prod_id, 'size', ARRAY(SELECT jsonb_array_elements_text(sizes)), NOW(), NOW()
      FROM products WHERE id = prod_id AND sizes IS NOT NULL AND sizes::text <> '[]';
    END IF;
  END LOOP;

  RAISE NOTICE 'seed-catalog: categories, brands, and catalog products ready';
END $$;
