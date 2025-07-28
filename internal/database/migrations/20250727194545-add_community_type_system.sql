
-- +migrate Up
-- Add community type system columns to subforums table
ALTER TABLE subforums ADD COLUMN community_type VARCHAR(10) NOT NULL DEFAULT 't';
ALTER TABLE subforums ADD COLUMN governance_style VARCHAR(20) NOT NULL DEFAULT 'democratic';
ALTER TABLE subforums ADD COLUMN owner_pseudonym_id VARCHAR(50);

-- Add index for community_type for efficient filtering
CREATE INDEX idx_subforums_community_type ON subforums(community_type);

-- Add index for owner_pseudonym_id for efficient lookups
CREATE INDEX idx_subforums_owner_pseudonym_id ON subforums(owner_pseudonym_id);

-- +migrate Down
-- Remove community type system columns from subforums table
DROP INDEX IF EXISTS idx_subforums_owner_pseudonym_id;
DROP INDEX IF EXISTS idx_subforums_community_type;
ALTER TABLE subforums DROP COLUMN IF EXISTS owner_pseudonym_id;
ALTER TABLE subforums DROP COLUMN IF EXISTS governance_style;
ALTER TABLE subforums DROP COLUMN IF EXISTS community_type;
