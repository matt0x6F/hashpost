-- +migrate Up
-- Migration: Add key_version field to conversation_keys table
-- This field tracks the version of the encryption key used for the conversation
-- to support key rotation and forward secrecy

-- Add key_version column to conversation_keys table
ALTER TABLE conversation_keys 
ADD COLUMN key_version INTEGER NOT NULL DEFAULT 1;

-- Create index on key_version for efficient lookups
CREATE INDEX idx_conversation_keys_key_version ON conversation_keys(key_version);

-- Add comment to document the field
COMMENT ON COLUMN conversation_keys.key_version IS 'Version of the encryption key used for this conversation';

-- +migrate Down
-- Remove the key_version column and index
DROP INDEX IF EXISTS idx_conversation_keys_key_version;
ALTER TABLE conversation_keys DROP COLUMN IF EXISTS key_version;
