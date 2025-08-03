
-- +migrate Up

-- Drop the old subforum_moderators table since we've migrated to the role_keys system
-- This table is no longer needed as moderation is now handled through role keys

-- Drop foreign key constraints first
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS fk_added_by_pseudonym;
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_pseudonym_id_fkey;
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_subforum_id_fkey;

-- Drop unique constraints
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_subforum_id_pseudonym_id_key;
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS unique_subforum_pseudonym;

-- Drop indexes
DROP INDEX IF EXISTS idx_moderators_pseudonym;
DROP INDEX IF EXISTS idx_moderators_subforum;

-- Drop the table
DROP TABLE IF EXISTS subforum_moderators;

-- Drop the sequence if it exists
DROP SEQUENCE IF EXISTS subforum_moderators_moderator_id_seq;

-- +migrate Down

-- Recreate the subforum_moderators table for rollback
CREATE SEQUENCE IF NOT EXISTS subforum_moderators_moderator_id_seq;

CREATE TABLE subforum_moderators (
    moderator_id BIGINT PRIMARY KEY DEFAULT nextval('subforum_moderators_moderator_id_seq'),
    subforum_id INTEGER NOT NULL,
    pseudonym_id CHARACTER VARYING(64) NOT NULL,
    role CHARACTER VARYING(20) NOT NULL DEFAULT 'moderator',
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    capabilities JSONB,
    added_by_pseudonym_id CHARACTER VARYING(64)
);

-- Recreate indexes
CREATE INDEX idx_moderators_pseudonym ON subforum_moderators(pseudonym_id);
CREATE INDEX idx_moderators_subforum ON subforum_moderators(subforum_id);
CREATE UNIQUE INDEX subforum_moderators_subforum_id_pseudonym_id_key ON subforum_moderators(subforum_id, pseudonym_id);
CREATE UNIQUE INDEX unique_subforum_pseudonym ON subforum_moderators(subforum_id, pseudonym_id);

-- Recreate foreign key constraints
ALTER TABLE subforum_moderators 
    ADD CONSTRAINT fk_added_by_pseudonym 
    FOREIGN KEY (added_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

ALTER TABLE subforum_moderators 
    ADD CONSTRAINT subforum_moderators_pseudonym_id_fkey 
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

ALTER TABLE subforum_moderators 
    ADD CONSTRAINT subforum_moderators_subforum_id_fkey 
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE;
