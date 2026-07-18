-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS privacy_rules (
    id                BIGSERIAL PRIMARY KEY,
    name              VARCHAR(255) NOT NULL,
    key               VARCHAR(100) NOT NULL,
    provider          VARCHAR(50)  NOT NULL,
    content_markdown  TEXT         NOT NULL DEFAULT '',
    summary           TEXT,
    version           INT          NOT NULL DEFAULT 1,
    locale            VARCHAR(10)  NOT NULL DEFAULT 'en',
    status            VARCHAR(50)  NOT NULL DEFAULT 'draft',
    workflow_state_id BIGINT       REFERENCES workflow_states(id) ON DELETE SET NULL,
    effective_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_privacy_rules_provider CHECK (
        provider IN ('platform', 'stripe', 'paypal', 'wallet', 'gift_card', 'shipping', 'ai', 'all')
    ),
    CONSTRAINT chk_privacy_rules_status CHECK (
        status IN ('draft', 'active', 'inactive', 'archived')
    ),
    CONSTRAINT uq_privacy_rules_key_locale UNIQUE (key, locale)
);

CREATE INDEX IF NOT EXISTS idx_privacy_rules_provider ON privacy_rules(provider);
CREATE INDEX IF NOT EXISTS idx_privacy_rules_status ON privacy_rules(status);
CREATE INDEX IF NOT EXISTS idx_privacy_rules_workflow_state_id ON privacy_rules(workflow_state_id);
CREATE INDEX IF NOT EXISTS idx_privacy_rules_locale ON privacy_rules(locale);

INSERT INTO workflows (key, name, description, entity_type) VALUES
    ('privacy_rule', 'Privacy Rule Lifecycle', 'Privacy / legal rule publishing lifecycle', 'privacy_rule')
ON CONFLICT (key) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('draft',    'Draft',    '#9CA3AF', TRUE,  FALSE, 1),
    ('active',   'Active',   '#10B981', FALSE, FALSE, 2),
    ('inactive', 'Inactive', '#6B7280', FALSE, FALSE, 3),
    ('archived', 'Archived', '#374151', FALSE, TRUE,  4)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'privacy_rule'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('draft',    'active',   'activate',   'Activate',   'admin', NULL, NULL),
    ('active',   'inactive', 'deactivate', 'Deactivate', 'admin', NULL, NULL),
    ('inactive', 'active',   'reactivate', 'Reactivate', 'admin', NULL, NULL),
    ('active',   'archived', 'archive',    'Archive',    'admin', NULL, NULL),
    ('inactive', 'archived', 'archive',    'Archive',    'admin', NULL, NULL),
    ('draft',    'archived', 'archive',    'Archive',    'admin', NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'privacy_rule'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM workflow_transition_logs WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'privacy_rule');
DELETE FROM workflow_transitions     WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'privacy_rule');
DELETE FROM workflow_states          WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'privacy_rule');
DELETE FROM workflows WHERE key = 'privacy_rule';
DROP TABLE IF EXISTS privacy_rules;
-- +goose StatementEnd
