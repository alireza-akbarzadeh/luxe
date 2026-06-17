-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workflows (
    id          BIGSERIAL    PRIMARY KEY,
    key         TEXT         NOT NULL UNIQUE,
    name        TEXT         NOT NULL,
    description TEXT,
    entity_type TEXT         NOT NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS workflow_states (
    id          BIGSERIAL    PRIMARY KEY,
    workflow_id BIGINT       NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    code        TEXT         NOT NULL,
    name        TEXT         NOT NULL,
    color       TEXT         NOT NULL DEFAULT '#6B7280',
    text_color  TEXT         NOT NULL DEFAULT '#FFFFFF',
    description TEXT,
    is_initial  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_final    BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (workflow_id, code)
);
CREATE INDEX IF NOT EXISTS idx_workflow_states_workflow ON workflow_states(workflow_id);

CREATE TABLE IF NOT EXISTS workflow_transitions (
    id            BIGSERIAL    PRIMARY KEY,
    workflow_id   BIGINT       NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    from_state_id BIGINT       REFERENCES workflow_states(id) ON DELETE CASCADE, -- NULL = wildcard "any state"
    to_state_id   BIGINT       NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    event         TEXT         NOT NULL,
    name          TEXT         NOT NULL,
    required_role TEXT,
    guard_key     TEXT,
    hook_key      TEXT,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    sort_order    INT          NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_workflow_transitions_workflow ON workflow_transitions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_transitions_from ON workflow_transitions(from_state_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_wf_trans_from_event ON workflow_transitions(workflow_id, from_state_id, event) WHERE from_state_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_wf_trans_any_event  ON workflow_transitions(workflow_id, event) WHERE from_state_id IS NULL;

CREATE TABLE IF NOT EXISTS workflow_transition_logs (
    id            BIGSERIAL    PRIMARY KEY,
    workflow_id   BIGINT       NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    entity_type   TEXT         NOT NULL,
    entity_id     BIGINT       NOT NULL,
    from_state_id BIGINT       REFERENCES workflow_states(id) ON DELETE SET NULL,
    to_state_id   BIGINT       REFERENCES workflow_states(id) ON DELETE SET NULL,
    transition_id BIGINT       REFERENCES workflow_transitions(id) ON DELETE SET NULL,
    event         TEXT         NOT NULL,
    user_id       BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    note          TEXT,
    success       BOOLEAN      NOT NULL DEFAULT TRUE,
    error_msg     TEXT,
    metadata      JSONB,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_wf_logs_entity  ON workflow_transition_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_wf_logs_created ON workflow_transition_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wf_logs_user    ON workflow_transition_logs(user_id);

CREATE TABLE IF NOT EXISTS returns (
    id                BIGSERIAL     PRIMARY KEY,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    order_id          BIGINT        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id           BIGINT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason            TEXT,
    status            TEXT          NOT NULL DEFAULT 'requested',
    refund_amount     DECIMAL(10,2) NOT NULL DEFAULT 0,
    workflow_state_id BIGINT        REFERENCES workflow_states(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_returns_order ON returns(order_id);
CREATE INDEX IF NOT EXISTS idx_returns_user  ON returns(user_id);

-- Add workflow_state_id pointers to existing entities. The engine becomes the
-- authority; the legacy status column is kept as an engine-written mirror.
ALTER TABLE orders    ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;
ALTER TABLE products  ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;
ALTER TABLE shipments ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;
ALTER TABLE users     ADD COLUMN IF NOT EXISTS workflow_state_id BIGINT REFERENCES workflow_states(id) ON DELETE SET NULL;

-- The product status CHECK forbids the new workflow state codes; drop it so the
-- status mirror can hold values like 'draft', 'under_review', 'published', etc.
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_status_check;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE orders    DROP COLUMN IF EXISTS workflow_state_id;
ALTER TABLE products  DROP COLUMN IF EXISTS workflow_state_id;
ALTER TABLE shipments DROP COLUMN IF EXISTS workflow_state_id;
ALTER TABLE users     DROP COLUMN IF EXISTS workflow_state_id;

DROP TABLE IF EXISTS returns;
DROP TABLE IF EXISTS workflow_transition_logs;
DROP TABLE IF EXISTS workflow_transitions;
DROP TABLE IF EXISTS workflow_states;
DROP TABLE IF EXISTS workflows;

ALTER TABLE products
    ADD CONSTRAINT products_status_check CHECK (status IN ('active', 'inactive', 'archived'));
-- +goose StatementEnd
