-- Fix schema issues for Bob model tests
-- 1. Fix owner_pseudonym_id column size to match pseudonym_id
-- 2. Remove duplicate constraint in identity_mappings

-- +migrate Up
-- Fix owner_pseudonym_id column size from VARCHAR(50) to VARCHAR(64)
ALTER TABLE subforums ALTER COLUMN owner_pseudonym_id TYPE VARCHAR(64);

-- Remove duplicate constraint (keep the one with the longer name)
ALTER TABLE identity_mappings DROP CONSTRAINT IF EXISTS unique_fingerprint_pseudonym_scope;

-- +migrate Down
-- Revert owner_pseudonym_id column size back to VARCHAR(50)
ALTER TABLE subforums ALTER COLUMN owner_pseudonym_id TYPE VARCHAR(50);

-- Re-add the duplicate constraint (this is the down migration)
ALTER TABLE identity_mappings ADD CONSTRAINT unique_fingerprint_pseudonym_scope UNIQUE (fingerprint, pseudonym_id, key_scope);
