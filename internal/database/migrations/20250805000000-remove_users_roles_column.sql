-- +migrate Up

-- Remove the roles column from users table since it's not used
-- All permissions are now managed through role_keys table
ALTER TABLE users DROP COLUMN IF EXISTS roles;

-- +migrate Down

-- Add back the roles column to users table
ALTER TABLE users ADD COLUMN roles JSONB DEFAULT '["user"]'; 