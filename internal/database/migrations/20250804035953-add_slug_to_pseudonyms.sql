
-- +migrate Up

-- Add slug field to pseudonyms table
ALTER TABLE pseudonyms ADD COLUMN slug VARCHAR(50);

-- Create unique index on slug
CREATE UNIQUE INDEX idx_pseudonyms_slug ON pseudonyms(slug);

-- Create index for slug lookups
CREATE INDEX idx_pseudonyms_slug_lookup ON pseudonyms(slug) WHERE slug IS NOT NULL;

-- +migrate Down

-- Remove indexes
DROP INDEX IF EXISTS idx_pseudonyms_slug_lookup;
DROP INDEX IF EXISTS idx_pseudonyms_slug;

-- Remove slug column
ALTER TABLE pseudonyms DROP COLUMN IF EXISTS slug;
