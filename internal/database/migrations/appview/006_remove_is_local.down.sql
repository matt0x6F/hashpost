-- Rollback: Add back is_local field
ALTER TABLE appview_users ADD COLUMN is_local BOOLEAN DEFAULT TRUE;

-- Update is_local based on pds_source
UPDATE appview_users 
SET is_local = (pds_source = 'http://hashpost-pds:8080' OR pds_source IS NULL);

-- Recreate the index
CREATE INDEX idx_appview_users_is_local ON appview_users(is_local);
