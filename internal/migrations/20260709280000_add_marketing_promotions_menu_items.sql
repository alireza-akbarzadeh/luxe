-- +goose Up
-- +goose StatementBegin
-- Promotions + email marketing hub links under Marketing; sales analytics under Reports.

DELETE FROM menu_items WHERE href = '/dashboard/marketing/newsletters';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Promotions', NULL, 'Flame', NULL, 2
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
WHERE g.name = 'Marketing'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.label = 'Promotions' AND mi.parent_id = p.id
  );

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, promo.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items root ON root.label = 'Marketing' AND root.parent_id IS NULL
JOIN menu_items promo ON promo.label = 'Promotions' AND promo.parent_id = root.id
CROSS JOIN (VALUES
    ('Overview', '/dashboard/promotions', 'Flame', 1),
    ('Flash sales', '/dashboard/promotions/flash-sales', 'Bolt', 2),
    ('Banners', '/dashboard/promotions/banners', 'Photo', 3),
    ('Campaigns', '/dashboard/promotions/campaigns', 'Speakerphone', 4)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Marketing'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Email marketing', NULL, 'Mail', NULL, 3
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
WHERE g.name = 'Marketing'
  AND NOT EXISTS (
    SELECT 1 FROM menu_items mi WHERE mi.label = 'Email marketing' AND mi.parent_id = p.id
  );

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, email.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items root ON root.label = 'Marketing' AND root.parent_id IS NULL
JOIN menu_items email ON email.label = 'Email marketing' AND email.parent_id = root.id
CROSS JOIN (VALUES
    ('Overview', '/dashboard/marketing', 'Mail', 1),
    ('Subscribers', '/dashboard/marketing/subscribers', 'Users', 2),
    ('Templates', '/dashboard/marketing/templates', 'Template', 3),
    ('Campaigns', '/dashboard/marketing/campaigns', 'Send', 4)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Marketing'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = v.href);

UPDATE menu_items SET display_order = 1 WHERE href = '/dashboard/discounts';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/notifications';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Sales analytics', '/dashboard/analytics', 'ChartBar', NULL, 1
FROM menu_groups g
JOIN menu_items p ON p.label = 'Reports' AND p.parent_id IS NULL
WHERE g.name = 'Reports'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/analytics');

UPDATE menu_items SET display_order = 2 WHERE href = '/dashboard/reports/revenue'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/analytics');
UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/reports/traffic'
  AND EXISTS (SELECT 1 FROM menu_items WHERE href = '/dashboard/analytics');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href IN (
    '/dashboard/promotions',
    '/dashboard/promotions/flash-sales',
    '/dashboard/promotions/banners',
    '/dashboard/promotions/campaigns',
    '/dashboard/marketing',
    '/dashboard/marketing/subscribers',
    '/dashboard/marketing/templates',
    '/dashboard/marketing/campaigns',
    '/dashboard/analytics'
);

DELETE FROM menu_items WHERE label IN ('Promotions', 'Email marketing') AND href IS NULL;

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Newsletters', '/dashboard/marketing/newsletters', 'Mail', NULL, 2
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
WHERE g.name = 'Marketing'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/marketing/newsletters');

UPDATE menu_items SET display_order = 1 WHERE href = '/dashboard/reports/revenue';
UPDATE menu_items SET display_order = 2 WHERE href = '/dashboard/reports/traffic';
UPDATE menu_items SET display_order = 1 WHERE href = '/dashboard/discounts';
UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/notifications';
-- +goose StatementEnd
