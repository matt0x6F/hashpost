-- Remove prefix_type column from appview_subforums table
-- This migration removes the prefix_type field

DROP INDEX IF EXISTS idx_appview_subforums_prefix_type;
ALTER TABLE appview_subforums DROP COLUMN IF EXISTS prefix_type;
