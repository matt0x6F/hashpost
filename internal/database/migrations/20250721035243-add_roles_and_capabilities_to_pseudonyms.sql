-- +migrate Up

-- Add roles and capabilities fields to pseudonyms table
ALTER TABLE pseudonyms ADD COLUMN roles JSONB DEFAULT '["user"]';
ALTER TABLE pseudonyms ADD COLUMN capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]';

-- Remove capabilities field from users table
ALTER TABLE users DROP COLUMN capabilities;

-- +migrate Down

-- Add capabilities field back to users table
ALTER TABLE users ADD COLUMN capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]';

-- Remove roles and capabilities fields from pseudonyms table
ALTER TABLE pseudonyms DROP COLUMN roles;
ALTER TABLE pseudonyms DROP COLUMN capabilities;
