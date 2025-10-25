-- Rollback Lightweight External User Records

DROP INDEX IF EXISTS idx_appview_users_is_local;
DROP INDEX IF EXISTS idx_appview_users_pds_source;

ALTER TABLE appview_users DROP COLUMN IF EXISTS last_seen_at;
ALTER TABLE appview_users DROP COLUMN IF EXISTS is_local;
ALTER TABLE appview_users DROP COLUMN IF EXISTS pds_source;
