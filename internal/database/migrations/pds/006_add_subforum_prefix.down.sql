-- Remove prefix_type column from subforums table
-- This migration removes the prefix_type field

DROP INDEX IF EXISTS idx_subforums_prefix_type;
ALTER TABLE subforums DROP COLUMN IF EXISTS prefix_type;
