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
    ARRAY['https://images.unsplash.com/photo-1587836374828-4db968944bda?w=900'],
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
  )
  ON CONFLICT (slug) DO NOTHING;

  -- ── Product attributes (color / size pickers) ─────────────────────────────
  FOR prod_id IN
    SELECT id FROM products WHERE slug IN (
      'arielle-silk-midi-dress', 'cloud-cashmere-crew', 'milan-wool-coat',
      'verona-structured-leather-tote', 'arcadia-suede-ankle-boots',
      'classic-oxford-shirt', 'navy-slim-wool-blazer', 'everyday-stretch-chino',
      'urban-chelsea-boot', 'stellar-automatic-38', 'heritage-chronograph-42',
      'essential-organic-cotton-tee', 'merino-roll-neck-sweater'
    )
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
