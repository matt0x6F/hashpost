
-- +migrate Up
-- Remove the redundant rules_text column since we now use the structured subforum_rules JSONB field
ALTER TABLE subforums DROP COLUMN IF EXISTS rules_text;

-- +migrate Down
-- Add back the rules_text column for rollback
ALTER TABLE subforums ADD COLUMN rules_text TEXT;
