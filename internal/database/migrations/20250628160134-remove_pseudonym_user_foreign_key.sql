-- +migrate Up
-- Migration to remove direct user_id foreign key from pseudonyms table
-- This enforces privacy by requiring IBE correlation for user-pseudonym relationships

-- Step 1: Populate user_id in identity_mappings from existing pseudonym-user relationships
-- Do this BEFORE we drop the user_id column from pseudonyms
UPDATE identity_mappings SET user_id = (
    SELECT p.user_id FROM pseudonyms p 
    WHERE p.pseudonym_id = identity_mappings.pseudonym_id
) WHERE user_id IS NULL;

-- Step 2: Drop foreign key constraint and index from pseudonyms
ALTER TABLE pseudonyms DROP CONSTRAINT IF EXISTS pseudonyms_user_id_fkey;
DROP INDEX IF EXISTS idx_pseudonyms_user;

-- Step 3: Remove the user_id column entirely from pseudonyms
ALTER TABLE pseudonyms DROP COLUMN user_id;

-- +migrate Down
-- Migration to restore direct user_id foreign key in pseudonyms table

-- Step 1: Add user_id column back to pseudonyms
ALTER TABLE pseudonyms ADD COLUMN user_id BIGINT;

-- Step 2: Populate user_id from identity_mappings
UPDATE pseudonyms SET user_id = (
    SELECT im.user_id FROM identity_mappings im 
    WHERE im.pseudonym_id = pseudonyms.pseudonym_id
);

-- Step 3: Make user_id NOT NULL and add foreign key constraint
ALTER TABLE pseudonyms ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE pseudonyms ADD CONSTRAINT pseudonyms_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- Step 4: Recreate index
CREATE INDEX idx_pseudonyms_user ON pseudonyms(user_id); 