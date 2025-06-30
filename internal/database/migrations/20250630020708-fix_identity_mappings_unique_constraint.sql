-- +migrate Up

-- Drop the old unique index that was created by the initial schema
-- This index was created as: CREATE UNIQUE INDEX unique_fingerprint_pseudonym ON identity_mappings (fingerprint, pseudonym_id);
DROP INDEX IF EXISTS unique_fingerprint_pseudonym;

-- Add the new unique constraint that includes key_scope
-- This constraint allows multiple mappings per fingerprint/pseudonym combination as long as they have different key_scopes
ALTER TABLE identity_mappings 
ADD CONSTRAINT unique_fingerprint_pseudonym_key_scope 
UNIQUE (fingerprint, pseudonym_id, key_scope);

-- +migrate Down

-- Drop the new unique constraint
ALTER TABLE identity_mappings DROP CONSTRAINT IF EXISTS unique_fingerprint_pseudonym_key_scope;

-- Recreate the old unique index
CREATE UNIQUE INDEX unique_fingerprint_pseudonym ON identity_mappings (fingerprint, pseudonym_id);
