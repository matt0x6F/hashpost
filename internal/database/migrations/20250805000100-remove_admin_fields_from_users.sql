-- +migrate Up
-- Migration: Remove deprecated admin fields from users table
-- This migration removes the admin_username, admin_password_hash, and admin_scope fields
-- from the users table as they are no longer used in the current permission system.

-- Remove admin_username column
ALTER TABLE users DROP COLUMN IF EXISTS admin_username;

-- Remove admin_password_hash column  
ALTER TABLE users DROP COLUMN IF EXISTS admin_password_hash;

-- Remove admin_scope column
ALTER TABLE users DROP COLUMN IF EXISTS admin_scope;

-- Drop any related indexes or constraints
DROP INDEX IF EXISTS idx_users_admin_username;

-- +migrate Down
-- Migration: Restore deprecated admin fields to users table
-- This migration restores the admin fields for rollback purposes.

-- Add admin_username column back
ALTER TABLE users ADD COLUMN admin_username VARCHAR(100) UNIQUE;

-- Add admin_password_hash column back
ALTER TABLE users ADD COLUMN admin_password_hash VARCHAR(255);

-- Add admin_scope column back
ALTER TABLE users ADD COLUMN admin_scope VARCHAR(100);

-- Recreate index
CREATE INDEX idx_users_admin_username ON users(admin_username); 