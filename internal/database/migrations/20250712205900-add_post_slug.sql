-- +migrate Up

-- Add slug column to posts table
ALTER TABLE posts ADD COLUMN slug VARCHAR(255);

-- Create index on slug for efficient lookups
CREATE INDEX idx_posts_slug ON posts(slug);

-- Add unique constraint on slug within each subforum
CREATE UNIQUE INDEX idx_posts_subforum_slug ON posts(subforum_id, slug);

-- +migrate Down

-- Drop indexes
DROP INDEX IF EXISTS idx_posts_subforum_slug;
DROP INDEX IF EXISTS idx_posts_slug;

-- Drop slug column
ALTER TABLE posts DROP COLUMN IF EXISTS slug;
