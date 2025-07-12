-- +migrate Up
-- Migration: Move subforum_moderators to IBE-based system (remove user_id, added_by_user_id; add added_by_pseudonym_id)

-- 1. Remove unique constraint on (subforum_id, user_id)
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_subforum_id_user_id_key;

-- 2. Drop the user_id column
ALTER TABLE subforum_moderators DROP COLUMN IF EXISTS user_id;

-- 3. Add added_by_pseudonym_id
ALTER TABLE subforum_moderators ADD COLUMN added_by_pseudonym_id VARCHAR(64);

-- 4. Drop the old added_by_user_id column
ALTER TABLE subforum_moderators DROP COLUMN IF EXISTS added_by_user_id;

-- 5. Add foreign key for added_by_pseudonym_id
ALTER TABLE subforum_moderators ADD CONSTRAINT fk_added_by_pseudonym FOREIGN KEY (added_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- 6. Ensure unique constraint on (subforum_id, pseudonym_id)
ALTER TABLE subforum_moderators ADD CONSTRAINT unique_subforum_pseudonym UNIQUE (subforum_id, pseudonym_id);

-- +migrate Down
-- Rollback: Restore user_id and added_by_user_id, remove added_by_pseudonym_id

-- 1. Remove unique constraint on (subforum_id, pseudonym_id)
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS unique_subforum_pseudonym;

-- 2. Drop foreign key and column for added_by_pseudonym_id
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS fk_added_by_pseudonym;
ALTER TABLE subforum_moderators DROP COLUMN IF EXISTS added_by_pseudonym_id;

-- 3. Add user_id column back
ALTER TABLE subforum_moderators ADD COLUMN user_id BIGINT;

-- 4. Add added_by_user_id column back
ALTER TABLE subforum_moderators ADD COLUMN added_by_user_id BIGINT;

-- 5. Restore unique constraint on (subforum_id, user_id)
ALTER TABLE subforum_moderators ADD CONSTRAINT subforum_moderators_subforum_id_user_id_key UNIQUE (subforum_id, user_id);

-- 6. Add foreign keys
ALTER TABLE subforum_moderators ADD FOREIGN KEY (user_id) REFERENCES users(user_id);
ALTER TABLE subforum_moderators ADD FOREIGN KEY (added_by_user_id) REFERENCES users(user_id); 