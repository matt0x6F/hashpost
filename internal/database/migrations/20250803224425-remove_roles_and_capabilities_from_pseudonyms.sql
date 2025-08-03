
-- +migrate Up

-- Remove roles and capabilities columns from pseudonyms table
-- These are now handled by the role_keys table
ALTER TABLE pseudonyms DROP COLUMN roles;
ALTER TABLE pseudonyms DROP COLUMN capabilities;

-- +migrate Down

-- Add back the roles and capabilities columns to pseudonyms table
ALTER TABLE pseudonyms ADD COLUMN roles JSONB DEFAULT '["user"]';
ALTER TABLE pseudonyms ADD COLUMN capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]';
