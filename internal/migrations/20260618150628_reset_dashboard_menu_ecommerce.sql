-- +goose Up
-- +goose StatementBegin
-- Replace legacy streaming-platform sidebar seed with Luxe e-commerce admin routes.

DELETE FROM menu_items;
DELETE FROM menu_groups;

INSERT INTO menu_groups (name, display_order) VALUES
('Overview', 1),
('Users & Access', 2),
('Catalog', 3),
('Orders & Fulfillment', 4),
('Marketing', 5),
('Reports', 6),
('Administration', 7);

-- Overview (leaf items)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
CROSS JOIN (VALUES
    ('Dashboard', '/dashboard', 'LayoutDashboard', 1),
    ('Live Sales Feed', '/dashboard/live', 'Activity', 2)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Overview';

-- Users & Access (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Users & Access', NULL, 'Users', NULL, 1
FROM menu_groups g WHERE g.name = 'Users & Access';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Users & Access' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Users', '/dashboard/users', 'Users', 1),
    ('Roles & Permissions', '/dashboard/roles', 'Shield', 2),
    ('Audit Logs', '/dashboard/audit-logs', 'FileText', 3)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Users & Access';

-- Catalog (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Catalog', NULL, 'Package', NULL, 1
FROM menu_groups g WHERE g.name = 'Catalog';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Catalog' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Products', '/dashboard/products', 'Package', 1),
    ('Categories', '/dashboard/categories', 'Tags', 2),
    ('Brands', '/dashboard/brands', 'BuildingStore', 3),
    ('Inventory', '/dashboard/inventory', 'Boxes', 4)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Catalog';

-- Orders & Fulfillment (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Orders & Fulfillment', NULL, 'Truck', NULL, 1
FROM menu_groups g WHERE g.name = 'Orders & Fulfillment';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Orders & Fulfillment' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Orders', '/dashboard/orders', 'ShoppingCart', 1),
    ('Shipments', '/dashboard/shipments', 'Truck', 2),
    ('Returns', '/dashboard/returns', 'ReceiptRefund', 3),
    ('Invoices', '/dashboard/invoices', 'Receipt', 4)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Orders & Fulfillment';

-- Marketing (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Marketing', NULL, 'Bell', NULL, 1
FROM menu_groups g WHERE g.name = 'Marketing';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Marketing' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Discounts', '/dashboard/discounts', 'Ticket', 1),
    ('Newsletters', '/dashboard/marketing/newsletters', 'Mail', 2),
    ('Notifications', '/dashboard/notifications', 'Bell', 3)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Marketing';

-- Reports (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Reports', NULL, 'BarChart3', NULL, 1
FROM menu_groups g WHERE g.name = 'Reports';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Reports' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Revenue', '/dashboard/reports/revenue', 'TrendingUp', 1),
    ('Traffic', '/dashboard/reports/traffic', 'DeviceAnalytics', 2)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Reports';

-- Administration (parent + children)
INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, NULL, 'Administration', NULL, 'Settings', NULL, 1
FROM menu_groups g WHERE g.name = 'Administration';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, v.label, v.href, v.icon, NULL, v.display_order
FROM menu_groups g
JOIN menu_items p ON p.label = 'Administration' AND p.parent_id IS NULL
CROSS JOIN (VALUES
    ('Suppliers', '/dashboard/suppliers', 'BuildingWarehouse', 1),
    ('Staff', '/dashboard/staff', 'UserShield', 2),
    ('Menus', '/dashboard/menus', 'Layout', 3),
    ('Workflows', '/dashboard/workflows', 'GitBranch', 4),
    ('Shipping', '/dashboard/settings/shipping', 'Truck', 5),
    ('Payment Gateways', '/dashboard/settings/gateways', 'CreditCard', 6),
    ('System', '/dashboard/settings/system', 'Server', 7)
) AS v(label, href, icon, display_order)
WHERE g.name = 'Administration';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items;
DELETE FROM menu_groups;
-- +goose StatementEnd
