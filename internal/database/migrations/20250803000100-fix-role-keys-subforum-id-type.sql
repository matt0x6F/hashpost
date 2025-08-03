-- +migrate Up
-- Fix type mismatch between subforums.subforum_id (integer) and role_keys.subforum_id (bigint)
-- Change role_keys.subforum_id to integer to match the referenced column

-- Step 1: Drop the foreign key constraint temporarily
ALTER TABLE role_keys DROP CONSTRAINT role_keys_subforum_id_fkey;

-- Step 2: Change the column type from bigint to integer
ALTER TABLE role_keys ALTER COLUMN subforum_id TYPE integer USING subforum_id::integer;

-- Step 3: Re-add the foreign key constraint
ALTER TABLE role_keys ADD CONSTRAINT role_keys_subforum_id_fkey 
FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id);

-- +migrate Down
-- Revert the type change
ALTER TABLE role_keys DROP CONSTRAINT role_keys_subforum_id_fkey;
ALTER TABLE role_keys ALTER COLUMN subforum_id TYPE bigint USING subforum_id::bigint;
ALTER TABLE role_keys ADD CONSTRAINT role_keys_subforum_id_fkey 
FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id); 