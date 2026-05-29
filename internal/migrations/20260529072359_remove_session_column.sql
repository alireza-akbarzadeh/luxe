-- +goose Up
-- +goose StatementBegin
-- First, delete any rows without a user_id (guest entries)
DELETE FROM compare_lists WHERE user_id IS NULL;

-- Drop the existing unique constraint that included session_id
ALTER TABLE compare_lists DROP CONSTRAINT IF EXISTS unique_user_or_session;

-- Drop the session_id column
ALTER TABLE compare_lists DROP COLUMN IF EXISTS session_id;

-- Ensure user_id is now NOT NULL (since we only support authenticated users)
ALTER TABLE compare_lists ALTER COLUMN user_id SET NOT NULL;

-- Add a new unique constraint on user_id alone
ALTER TABLE compare_lists ADD CONSTRAINT unique_user_compare UNIQUE (user_id);

-- Drop the index on session_id (if it exists)
DROP INDEX IF EXISTS idx_compare_lists_session_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Re-add session_id column (nullable)
ALTER TABLE compare_lists ADD COLUMN session_id VARCHAR(255);

-- Re-create the unique constraint on (user_id, session_id)
ALTER TABLE compare_lists ADD CONSTRAINT unique_user_or_session UNIQUE (user_id, session_id);

-- Allow user_id to be nullable again
ALTER TABLE compare_lists ALTER COLUMN user_id DROP NOT NULL;

-- Re-create the index on session_id
CREATE INDEX idx_compare_lists_session_id ON compare_lists(session_id);

-- (Note: data lost during the up migration cannot be restored automatically)
-- +goose StatementEnd