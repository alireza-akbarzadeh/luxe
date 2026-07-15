-- +goose Up
-- +goose StatementBegin

ALTER TABLE blog_posts
    ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_blog_posts_workflow_state_id ON blog_posts(workflow_state_id);

INSERT INTO workflows (key, name, description, entity_type) VALUES
    ('blog_post', 'Blog Post Lifecycle', 'Editorial draft, review, schedule, publish, and archive', 'blog_post')
ON CONFLICT (key) DO NOTHING;

INSERT INTO workflow_states (workflow_id, code, name, color, is_initial, is_final, sort_order)
SELECT w.id, v.code, v.name, v.color, v.is_initial, v.is_final, v.sort_order
FROM workflows w, (VALUES
    ('draft',     'Draft',     '#9CA3AF', TRUE,  FALSE, 1),
    ('in_review', 'In Review', '#F59E0B', FALSE, FALSE, 2),
    ('scheduled', 'Scheduled', '#3B82F6', FALSE, FALSE, 3),
    ('published', 'Published', '#10B981', FALSE, FALSE, 4),
    ('archived',  'Archived',  '#374151', FALSE, TRUE,  5)
) AS v(code, name, color, is_initial, is_final, sort_order)
WHERE w.key = 'blog_post'
ON CONFLICT (workflow_id, code) DO NOTHING;

INSERT INTO workflow_transitions (workflow_id, from_state_id, to_state_id, event, name, required_role, guard_key, hook_key)
SELECT w.id, fs.id, ts.id, t.event, t.name, t.required_role, t.guard_key, t.hook_key
FROM workflows w
JOIN (VALUES
    ('draft',     'in_review', 'submit_review',    'Submit for review', 'admin', NULL, NULL),
    ('in_review', 'scheduled', 'approve_schedule', 'Approve & schedule','admin', NULL, NULL),
    ('in_review', 'published', 'publish',          'Publish',           'admin', NULL, 'blog_post_published'),
    ('draft',     'published', 'publish',          'Publish',           'admin', NULL, 'blog_post_published'),
    ('scheduled', 'published', 'publish',          'Publish',           'admin', NULL, 'blog_post_published'),
    ('draft',     'scheduled', 'schedule',         'Schedule',          'admin', NULL, NULL),
    ('published', 'draft',     'unpublish',        'Unpublish',         'admin', NULL, NULL),
    ('published', 'archived',  'archive',          'Archive',           'admin', NULL, NULL),
    ('scheduled', 'archived',  'archive',          'Archive',           'admin', NULL, NULL),
    ('in_review', 'archived',  'archive',          'Archive',           'admin', NULL, NULL),
    ('draft',     'archived',  'archive',          'Archive',           'admin', NULL, NULL),
    ('archived',  'draft',     'restore',          'Restore to draft',  'admin', NULL, NULL)
) AS t(from_code, to_code, event, name, required_role, guard_key, hook_key) ON TRUE
JOIN workflow_states fs ON fs.workflow_id = w.id AND fs.code = t.from_code
JOIN workflow_states ts ON ts.workflow_id = w.id AND ts.code = t.to_code
WHERE w.key = 'blog_post'
ON CONFLICT DO NOTHING;

UPDATE blog_posts p SET workflow_state_id = s.id
FROM workflow_states s
JOIN workflows w ON w.id = s.workflow_id AND w.key = 'blog_post'
WHERE p.workflow_state_id IS NULL
  AND s.code = CASE p.status
      WHEN 'draft'     THEN 'draft'
      WHEN 'in_review' THEN 'in_review'
      WHEN 'scheduled' THEN 'scheduled'
      WHEN 'published' THEN 'published'
      WHEN 'archived'  THEN 'archived'
      ELSE 'draft'
  END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE blog_posts SET workflow_state_id = NULL;
DELETE FROM workflow_transition_logs WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'blog_post');
DELETE FROM workflow_transitions     WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'blog_post');
DELETE FROM workflow_states          WHERE workflow_id IN (SELECT id FROM workflows WHERE key = 'blog_post');
DELETE FROM workflows WHERE key = 'blog_post';
ALTER TABLE blog_posts DROP COLUMN IF EXISTS workflow_state_id;
-- +goose StatementEnd
