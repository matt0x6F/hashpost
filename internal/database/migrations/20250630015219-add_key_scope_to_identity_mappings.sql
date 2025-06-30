-- +migrate Up

-- Add key_scope column to identity_mappings table
ALTER TABLE identity_mappings ADD COLUMN key_scope VARCHAR(50) NOT NULL DEFAULT 'correlation';

-- Add new unique constraint that includes key_scope
ALTER TABLE identity_mappings ADD CONSTRAINT unique_fingerprint_pseudonym_scope 
    UNIQUE (fingerprint, pseudonym_id, key_scope);

-- Update existing records to have the default key_scope
UPDATE identity_mappings SET key_scope = 'correlation' WHERE key_scope IS NULL OR key_scope = '';

-- +migrate Down
