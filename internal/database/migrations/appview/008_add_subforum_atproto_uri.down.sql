-- Remove atproto_uri field from appview_subforums table
DROP INDEX IF EXISTS idx_appview_subforums_atproto_uri;
ALTER TABLE appview_subforums DROP COLUMN IF EXISTS atproto_uri;
