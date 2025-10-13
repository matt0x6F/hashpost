-- Remove password hash column from users table
-- This migration removes password hashing support

-- Drop the index first
DROP INDEX IF EXISTS idx_users_password_hash;

-- Remove the password hash column
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
