-- +migrate Up
-- Migration: Migrate role keys to pseudonym-based system with subforum scoping
-- This migration:
-- 1. Adds pseudonym_id and subforum_id columns to role_keys
-- 2. Migrates existing subforum moderators to role keys
-- 3. Renames permissions to capabilities in subforum_moderators
-- 4. Drops the subforum_moderators table after migration

-- Step 1: Add new columns to role_keys table (without foreign key constraints initially)
ALTER TABLE role_keys 
ADD COLUMN pseudonym_id CHARACTER VARYING(64),
ADD COLUMN subforum_id BIGINT;

-- Create indexes for efficient lookups
CREATE INDEX idx_role_keys_pseudonym ON role_keys(pseudonym_id);
CREATE INDEX idx_role_keys_subforum ON role_keys(subforum_id);
CREATE INDEX idx_role_keys_pseudonym_subforum ON role_keys(pseudonym_id, subforum_id);

-- Step 2: Rename permissions to capabilities in subforum_moderators
ALTER TABLE subforum_moderators RENAME COLUMN permissions TO capabilities;

-- Step 3: Change created_by column type to character varying before inserting string data
ALTER TABLE role_keys DROP CONSTRAINT role_keys_created_by_fkey;
ALTER TABLE role_keys ALTER COLUMN created_by TYPE CHARACTER VARYING(64);

-- Step 4: Migrate existing subforum moderators to role keys
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
)
SELECT 
    gen_random_uuid() as key_id,
    sm.role as role_name,
    'moderation' as scope,
    encode(sha256(concat(sm.pseudonym_id::text, sm.role, 'moderation')::bytea), 'hex')::bytea as key_data,
    1 as key_version,
    COALESCE(sm.capabilities, '[]'::jsonb) as capabilities,
    COALESCE(sm.added_at, NOW()) as created_at,
    NOW() + INTERVAL '1 year' as expires_at,
    true as is_active,
    COALESCE(sm.added_by_pseudonym_id, sm.pseudonym_id) as created_by,
    sm.pseudonym_id as pseudonym_id,
    sm.subforum_id as subforum_id
FROM subforum_moderators sm
WHERE sm.pseudonym_id IS NOT NULL;

-- Step 5: Create global role keys for platform admins
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
)
SELECT 
    gen_random_uuid() as key_id,
    'platform_admin' as role_name,
    'correlation' as scope,
    encode(sha256(concat(p.pseudonym_id::text, 'platform_admin', 'correlation')::bytea), 'hex')::bytea as key_data,
    1 as key_version,
    '["access_all_pseudonyms", "cross_user_correlation", "system_admin"]'::jsonb as capabilities,
    NOW() as created_at,
    NOW() + INTERVAL '1 year' as expires_at,
    true as is_active,
    p.pseudonym_id as created_by,
    p.pseudonym_id as pseudonym_id,
    NULL as subforum_id
FROM users u
JOIN identity_mappings im ON u.user_id = im.user_id
JOIN pseudonyms p ON im.pseudonym_id = p.pseudonym_id AND p.is_default = true
WHERE u.roles::text LIKE '%platform_admin%';

-- Step 6: Create global role keys for trust_safety
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
)
SELECT 
    gen_random_uuid() as key_id,
    'trust_safety' as role_name,
    'correlation' as scope,
    encode(sha256(concat(p.pseudonym_id::text, 'trust_safety', 'correlation')::bytea), 'hex')::bytea as key_data,
    1 as key_version,
    '["access_all_pseudonyms", "cross_user_correlation", "moderation"]'::jsonb as capabilities,
    NOW() as created_at,
    NOW() + INTERVAL '1 year' as expires_at,
    true as is_active,
    p.pseudonym_id as created_by,
    p.pseudonym_id as pseudonym_id,
    NULL as subforum_id
FROM users u
JOIN identity_mappings im ON u.user_id = im.user_id
JOIN pseudonyms p ON im.pseudonym_id = p.pseudonym_id AND p.is_default = true
WHERE u.roles::text LIKE '%trust_safety%';

-- Step 7: Create global role keys for legal_team
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
)
SELECT 
    gen_random_uuid() as key_id,
    'legal_team' as role_name,
    'correlation' as scope,
    encode(sha256(concat(p.pseudonym_id::text, 'legal_team', 'correlation')::bytea), 'hex')::bytea as key_data,
    1 as key_version,
    '["access_all_pseudonyms", "cross_user_correlation", "legal_compliance"]'::jsonb as capabilities,
    NOW() as created_at,
    NOW() + INTERVAL '1 year' as expires_at,
    true as is_active,
    p.pseudonym_id as created_by,
    p.pseudonym_id as pseudonym_id,
    NULL as subforum_id
FROM users u
JOIN identity_mappings im ON u.user_id = im.user_id
JOIN pseudonyms p ON im.pseudonym_id = p.pseudonym_id AND p.is_default = true
WHERE u.roles::text LIKE '%legal_team%';

-- Step 8: Create authentication role keys for all users
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
)
SELECT 
    gen_random_uuid() as key_id,
    'user' as role_name,
    'authentication' as scope,
    encode(sha256(concat(p.pseudonym_id::text, 'user', 'authentication')::bytea), 'hex')::bytea as key_data,
    1 as key_version,
    '["access_own_pseudonyms", "login", "session_management"]'::jsonb as capabilities,
    NOW() as created_at,
    NOW() + INTERVAL '1 year' as expires_at,
    true as is_active,
    p.pseudonym_id as created_by,
    p.pseudonym_id as pseudonym_id,
    NULL as subforum_id
FROM users u
JOIN identity_mappings im ON u.user_id = im.user_id
JOIN pseudonyms p ON im.pseudonym_id = p.pseudonym_id AND p.is_default = true
WHERE u.roles IS NOT NULL AND u.roles != '[]'::jsonb;

-- Step 9: Delete legacy rows with NULL pseudonym_id before setting NOT NULL constraint
DELETE FROM role_keys WHERE pseudonym_id IS NULL;

-- Step 10: Add foreign key constraints after data migration
ALTER TABLE role_keys ADD CONSTRAINT role_keys_pseudonym_id_fkey 
FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

ALTER TABLE role_keys ADD CONSTRAINT role_keys_subforum_id_fkey 
FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id);

-- Step 11: Make pseudonym_id NOT NULL after migration
ALTER TABLE role_keys ALTER COLUMN pseudonym_id SET NOT NULL;

-- Step 12: Add new foreign key to pseudonyms for created_by
ALTER TABLE role_keys ADD CONSTRAINT role_keys_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES pseudonyms(pseudonym_id);

-- Step 13: Drop the old subforum_moderators table
DROP TABLE subforum_moderators;

-- +migrate Down
-- Revert the migration by dropping the new columns and restoring the old structure
ALTER TABLE role_keys DROP COLUMN IF EXISTS pseudonym_id;
ALTER TABLE role_keys DROP COLUMN IF EXISTS subforum_id;

-- Drop the indexes
DROP INDEX IF EXISTS idx_role_keys_pseudonym;
DROP INDEX IF EXISTS idx_role_keys_subforum;
DROP INDEX IF EXISTS idx_role_keys_pseudonym_subforum;

-- Restore the old foreign key constraint
ALTER TABLE role_keys DROP CONSTRAINT IF EXISTS role_keys_created_by_fkey;
ALTER TABLE role_keys ALTER COLUMN created_by TYPE BIGINT;
ALTER TABLE role_keys ADD CONSTRAINT role_keys_created_by_fkey 
FOREIGN KEY (created_by) REFERENCES users(user_id);

-- Note: We cannot easily restore the subforum_moderators data since it was migrated
-- This would require a more complex rollback strategy 