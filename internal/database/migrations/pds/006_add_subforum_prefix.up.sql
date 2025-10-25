-- Add prefix_type column to subforums table
-- This migration adds a prefix_type field to standardize subforum slugs
-- with prefixes: h- (HashPost, admin-only), r- (regional), t- (topical)

ALTER TABLE subforums 
ADD COLUMN prefix_type VARCHAR(1) NOT NULL DEFAULT 't' 
CHECK (prefix_type IN ('h', 'r', 't'));

-- Add index for filtering by prefix type
CREATE INDEX idx_subforums_prefix_type ON subforums(prefix_type);

-- Update existing subforums to have 't' prefix (topical)
-- This is already the default, but being explicit
UPDATE subforums SET prefix_type = 't' WHERE prefix_type IS NULL;
