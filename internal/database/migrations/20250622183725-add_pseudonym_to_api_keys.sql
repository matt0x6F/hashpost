-- +migrate Up
-- Migration to add pseudonym_id to api_keys table
-- API keys should belong to a pseudonym as they represent programmatic access for that pseudonym

-- Add pseudonym_id column to api_keys table
ALTER TABLE api_keys ADD COLUMN pseudonym_id VARCHAR(64);

-- Add foreign key constraint
ALTER TABLE api_keys ADD CONSTRAINT fk_api_keys_pseudonym 
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Create index for better query performance
CREATE INDEX idx_api_keys_pseudonym ON api_keys(pseudonym_id);

-- Make pseudonym_id NOT NULL after adding the constraint
-- Note: This will fail if there are existing API keys, so we'll handle that in the down migration
-- ALTER TABLE api_keys ALTER COLUMN pseudonym_id SET NOT NULL;

-- +migrate Down
-- Rollback migration to remove pseudonym_id from api_keys table

-- Drop the index
DROP INDEX IF EXISTS idx_api_keys_pseudonym;

-- Drop the foreign key constraint
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS fk_api_keys_pseudonym;

-- Drop the column
ALTER TABLE api_keys DROP COLUMN IF EXISTS pseudonym_id;
