-- +migrate Up
-- Migration: Add unique constraint to pseudonym display names
-- This ensures that each display name is unique across all pseudonyms

-- Drop the existing non-unique index if it exists
DROP INDEX IF EXISTS idx_pseudonyms_display_name;

-- Add unique constraint on display_name (creates a unique index)
ALTER TABLE pseudonyms ADD CONSTRAINT unique_pseudonym_display_name UNIQUE (display_name);

-- +migrate Down
-- Rollback: Remove unique constraint from pseudonym display names

-- Drop the unique constraint
ALTER TABLE pseudonyms DROP CONSTRAINT IF EXISTS unique_pseudonym_display_name;

-- Recreate the non-unique index
CREATE INDEX idx_pseudonyms_display_name ON pseudonyms(display_name); 