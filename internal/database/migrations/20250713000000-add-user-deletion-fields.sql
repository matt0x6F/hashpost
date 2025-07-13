-- +migrate Up

-- Add pseudonym-based deletion fields to posts table
ALTER TABLE posts ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE posts ADD COLUMN deleted_by_pseudonym_id VARCHAR(64);
ALTER TABLE posts ADD COLUMN deleted_by_pseudonym_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE posts ADD COLUMN deleted_by_pseudonym_reason VARCHAR(100);
ALTER TABLE posts ADD CONSTRAINT fk_posts_deleted_by_pseudonym FOREIGN KEY (deleted_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Add pseudonym-based deletion fields to comments table
ALTER TABLE comments ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE comments ADD COLUMN deleted_by_pseudonym_id VARCHAR(64);
ALTER TABLE comments ADD COLUMN deleted_by_pseudonym_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE comments ADD COLUMN deleted_by_pseudonym_reason VARCHAR(100);
ALTER TABLE comments ADD CONSTRAINT fk_comments_deleted_by_pseudonym FOREIGN KEY (deleted_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- +migrate Down

ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk_posts_deleted_by_pseudonym;
ALTER TABLE posts DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_by_pseudonym_id;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_by_pseudonym_at;
ALTER TABLE posts DROP COLUMN IF EXISTS deleted_by_pseudonym_reason;

ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_deleted_by_pseudonym;
ALTER TABLE comments DROP COLUMN IF EXISTS is_deleted;
ALTER TABLE comments DROP COLUMN IF EXISTS deleted_by_pseudonym_id;
ALTER TABLE comments DROP COLUMN IF EXISTS deleted_by_pseudonym_at;
ALTER TABLE comments DROP COLUMN IF EXISTS deleted_by_pseudonym_reason; 