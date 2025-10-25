-- Lightweight External User Records
-- This migration adds support for external PDS users with minimal local storage

-- Add columns to track user source
ALTER TABLE appview_users ADD COLUMN pds_source VARCHAR(500);
ALTER TABLE appview_users ADD COLUMN is_local BOOLEAN DEFAULT TRUE;
ALTER TABLE appview_users ADD COLUMN last_seen_at TIMESTAMPTZ;

-- Index for external user lookups
CREATE INDEX idx_appview_users_pds_source ON appview_users(pds_source) WHERE pds_source IS NOT NULL;
CREATE INDEX idx_appview_users_is_local ON appview_users(is_local);

-- Add comments
COMMENT ON COLUMN appview_users.pds_source IS 'PDS endpoint where user data is stored (NULL for local users)';
COMMENT ON COLUMN appview_users.is_local IS 'Whether user is stored on HashPost PDS (true) or external PDS (false)';
COMMENT ON COLUMN appview_users.last_seen_at IS 'Last time user authenticated (for external users)';
