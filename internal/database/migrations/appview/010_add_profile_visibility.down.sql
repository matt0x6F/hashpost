-- Remove profile visibility setting from appview_users table

-- Drop the index first
DROP INDEX IF EXISTS idx_appview_users_profile_visibility;

-- Remove the column
ALTER TABLE appview_users DROP COLUMN IF EXISTS profile_visibility;

-- Drop the enum type
DROP TYPE IF EXISTS profile_visibility;
