-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_teams_slug ON teams(slug);

INSERT INTO permissions (key, module, description, created_at, updated_at)
SELECT v.key, v.module, v.description, NOW(), NOW()
FROM (VALUES
  ('teams.read', 'teams', 'View teams and members'),
  ('teams.write', 'teams', 'Manage teams and membership')
) AS v(key, module, description)
WHERE NOT EXISTS (SELECT 1 FROM permissions p WHERE p.key = v.key);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('teams.read', 'teams.write')
WHERE r.slug = 'admin'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Teams', '/dashboard/teams', 'UsersGroup', NULL, 3
FROM menu_groups g
JOIN menu_items p ON p.label = 'Users' AND p.parent_id IS NOT NULL
WHERE g.name = 'Users & Access'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/teams');

UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/access-control';
UPDATE menu_items SET display_order = 5 WHERE href = '/dashboard/roles';
UPDATE menu_items SET display_order = 6 WHERE href = '/dashboard/audit-logs';

DELETE FROM menu_items WHERE href = '/dashboard/access-control';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM menu_items WHERE href = '/dashboard/teams';

INSERT INTO menu_items (group_id, parent_id, label, href, icon, permission, display_order)
SELECT g.id, p.id, 'Access Control', '/dashboard/access-control', 'Users', NULL, 3
FROM menu_groups g
JOIN menu_items p ON p.label = 'Users' AND p.parent_id IS NOT NULL
WHERE g.name = 'Users & Access'
  AND NOT EXISTS (SELECT 1 FROM menu_items mi WHERE mi.href = '/dashboard/access-control');

UPDATE menu_items SET display_order = 3 WHERE href = '/dashboard/access-control';
UPDATE menu_items SET display_order = 4 WHERE href = '/dashboard/roles';
UPDATE menu_items SET display_order = 5 WHERE href = '/dashboard/audit-logs';

DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE key IN ('teams.read', 'teams.write'));

DELETE FROM permissions WHERE key IN ('teams.read', 'teams.write');

DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
