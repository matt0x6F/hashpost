-- +migrate Up

DROP INDEX IF EXISTS unique_fingerprint_pseudonym;

-- +migrate Down

-- Recreate the old unique index if needed:
-- CREATE UNIQUE INDEX unique_fingerprint_pseudonym ON identity_mappings (fingerprint, pseudonym_id);
