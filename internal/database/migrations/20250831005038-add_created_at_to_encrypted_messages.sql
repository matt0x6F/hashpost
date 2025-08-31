
-- +migrate Up
-- Add created_at field to encrypted_messages table for message timestamp tracking

ALTER TABLE encrypted_messages 
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Create index for timestamp-based queries
CREATE INDEX idx_encrypted_messages_created_at ON encrypted_messages(created_at);

-- +migrate Down
-- Remove created_at field from encrypted_messages table

DROP INDEX IF EXISTS idx_encrypted_messages_created_at;
ALTER TABLE encrypted_messages DROP COLUMN IF EXISTS created_at;
