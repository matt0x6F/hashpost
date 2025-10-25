-- Remove is_local field - all PDS servers are equal in AT Protocol
-- This migration removes the is_local field and updates pds_source for local users

-- Update local users to have proper pds_source (only if column exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'appview_users' AND column_name = 'is_local') THEN
        UPDATE appview_users 
        SET pds_source = 'http://hashpost-pds:8080' 
        WHERE is_local = true AND pds_source IS NULL;
    END IF;
END $$;

-- Remove the is_local column (only if it exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'appview_users' AND column_name = 'is_local') THEN
        ALTER TABLE appview_users DROP COLUMN is_local;
    END IF;
END $$;

-- Remove the index
DROP INDEX IF EXISTS idx_appview_users_is_local;
