-- Add profile visibility setting to appview_users table
-- This allows users to control who can view their profiles

-- Create enum type for profile visibility
CREATE TYPE profile_visibility AS ENUM ('public', 'authenticated', 'private');

-- Add profile_visibility column to appview_users table
ALTER TABLE appview_users 
ADD COLUMN profile_visibility profile_visibility NOT NULL DEFAULT 'public';

-- Add index for efficient querying by visibility
CREATE INDEX idx_appview_users_profile_visibility ON appview_users(profile_visibility);

-- Add comment explaining the visibility levels
COMMENT ON COLUMN appview_users.profile_visibility IS 'Controls who can view this user''s profile: public (anyone), authenticated (logged in users), private (only the user themselves)';
