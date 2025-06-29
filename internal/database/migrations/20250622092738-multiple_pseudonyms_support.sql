-- +migrate Up
-- Migration to support multiple pseudonyms per user
-- This migration transforms the existing 1:1 user:pseudonym relationship to 1:many

-- Step 1: Create the new pseudonyms table
CREATE TABLE pseudonyms (
    pseudonym_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    karma_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Profile metadata (optional)
    bio TEXT,
    avatar_url VARCHAR(255),
    website_url VARCHAR(255),
    
    -- Privacy settings
    show_karma BOOLEAN DEFAULT TRUE,
    allow_direct_messages BOOLEAN DEFAULT TRUE,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Step 2: Create indexes for the pseudonyms table
CREATE INDEX idx_pseudonyms_user ON pseudonyms(user_id);
CREATE INDEX idx_pseudonyms_display_name ON pseudonyms(display_name);
CREATE INDEX idx_pseudonyms_karma_score ON pseudonyms(karma_score);
CREATE INDEX idx_pseudonyms_created_at ON pseudonyms(created_at);
CREATE INDEX idx_pseudonyms_last_active ON pseudonyms(last_active_at);
CREATE INDEX idx_pseudonyms_active ON pseudonyms(is_active);

-- Step 3: Migrate existing user data to pseudonyms table
-- Create a default pseudonym for each existing user
INSERT INTO pseudonyms (
    pseudonym_id,
    user_id,
    display_name,
    karma_score,
    created_at,
    last_active_at,
    is_active,
    bio,
    avatar_url,
    website_url,
    show_karma,
    allow_direct_messages
)
SELECT 
    pseudonym_id,
    user_id,
    display_name,
    karma_score,
    created_at,
    last_active_at,
    is_active,
    bio,
    avatar_url,
    website_url,
    show_karma,
    allow_direct_messages
FROM users;

-- Step 4: Update identity_mappings table to include user_id
ALTER TABLE identity_mappings ADD COLUMN user_id BIGINT;
UPDATE identity_mappings SET user_id = (
    SELECT user_id FROM pseudonyms WHERE pseudonyms.pseudonym_id = identity_mappings.pseudonym_id
);
ALTER TABLE identity_mappings ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE identity_mappings ADD FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE identity_mappings ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;
CREATE INDEX idx_mappings_user ON identity_mappings(user_id);

-- Step 5: Update content tables to use pseudonym_id instead of user_id
-- Posts table
ALTER TABLE posts ADD COLUMN pseudonym_id VARCHAR(64);
UPDATE posts SET pseudonym_id = (
    SELECT pseudonym_id FROM pseudonyms WHERE pseudonyms.user_id = posts.user_id
);
ALTER TABLE posts ALTER COLUMN pseudonym_id SET NOT NULL;
ALTER TABLE posts DROP COLUMN user_id;
ALTER TABLE posts ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;
CREATE INDEX idx_posts_pseudonym ON posts(pseudonym_id);

-- Comments table
ALTER TABLE comments ADD COLUMN pseudonym_id VARCHAR(64);
UPDATE comments SET pseudonym_id = (
    SELECT pseudonym_id FROM pseudonyms WHERE pseudonyms.user_id = comments.user_id
);
ALTER TABLE comments ALTER COLUMN pseudonym_id SET NOT NULL;
ALTER TABLE comments DROP COLUMN user_id;
ALTER TABLE comments ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;
CREATE INDEX idx_comments_pseudonym ON comments(pseudonym_id);

-- Step 6: Drop foreign key constraints that reference users.pseudonym_id before we can drop the column
-- Drop constraints from subforum_subscriptions
ALTER TABLE subforum_subscriptions DROP CONSTRAINT IF EXISTS subforum_subscriptions_pseudonym_id_fkey;

-- Drop constraints from votes
ALTER TABLE votes DROP CONSTRAINT IF EXISTS votes_pseudonym_id_fkey;

-- Drop constraints from poll_votes
ALTER TABLE poll_votes DROP CONSTRAINT IF EXISTS poll_votes_pseudonym_id_fkey;

-- Drop constraints from user_blocks
ALTER TABLE user_blocks DROP CONSTRAINT IF EXISTS user_blocks_blocker_pseudonym_id_fkey;
ALTER TABLE user_blocks DROP CONSTRAINT IF EXISTS user_blocks_blocked_pseudonym_id_fkey;

-- Drop constraints from direct_messages
ALTER TABLE direct_messages DROP CONSTRAINT IF EXISTS direct_messages_sender_pseudonym_id_fkey;
ALTER TABLE direct_messages DROP CONSTRAINT IF EXISTS direct_messages_recipient_pseudonym_id_fkey;

-- Drop constraints from reports
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_reporter_pseudonym_id_fkey;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_reported_pseudonym_id_fkey;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_resolved_by_pseudonym_id_fkey;

-- Drop constraints from user_bans
ALTER TABLE user_bans DROP CONSTRAINT IF EXISTS user_bans_banned_by_pseudonym_id_fkey;

-- Drop constraints from moderation_actions
ALTER TABLE moderation_actions DROP CONSTRAINT IF EXISTS moderation_actions_moderator_pseudonym_id_fkey;

-- Drop constraints from subforum_moderators
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_pseudonym_id_fkey;

-- Drop constraints from posts (moderation fields)
ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_removed_by_pseudonym_id_fkey;

-- Drop constraints from comments (moderation fields)
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_removed_by_pseudonym_id_fkey;

-- Drop constraints from identity_mappings
ALTER TABLE identity_mappings DROP CONSTRAINT IF EXISTS identity_mappings_pseudonym_id_fkey;

-- Step 7: Update foreign key references to point to the new pseudonyms table
-- Update votes table
ALTER TABLE votes ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update poll_votes table
ALTER TABLE poll_votes ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update subforum_subscriptions table
ALTER TABLE subforum_subscriptions ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update user_blocks table
ALTER TABLE user_blocks ADD FOREIGN KEY (blocker_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;
ALTER TABLE user_blocks ADD FOREIGN KEY (blocked_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update direct_messages table
ALTER TABLE direct_messages ADD FOREIGN KEY (sender_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;
ALTER TABLE direct_messages ADD FOREIGN KEY (recipient_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update reports table
ALTER TABLE reports ADD FOREIGN KEY (reporter_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);
ALTER TABLE reports ADD FOREIGN KEY (reported_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);
ALTER TABLE reports ADD FOREIGN KEY (resolved_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Update user_bans table
ALTER TABLE user_bans ADD FOREIGN KEY (banned_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Update moderation_actions table
ALTER TABLE moderation_actions ADD FOREIGN KEY (moderator_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Update subforum_moderators table
ALTER TABLE subforum_moderators ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE;

-- Update correlation_audit table
ALTER TABLE correlation_audit ADD FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Update posts moderation fields
ALTER TABLE posts ADD FOREIGN KEY (removed_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Update comments moderation fields
ALTER TABLE comments ADD FOREIGN KEY (removed_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id);

-- Step 8: Clean up users table - remove pseudonym-specific columns
ALTER TABLE users DROP COLUMN pseudonym_id;
ALTER TABLE users DROP COLUMN display_name;
ALTER TABLE users DROP COLUMN karma_score;
ALTER TABLE users DROP COLUMN bio;
ALTER TABLE users DROP COLUMN avatar_url;
ALTER TABLE users DROP COLUMN website_url;
ALTER TABLE users DROP COLUMN show_karma;
ALTER TABLE users DROP COLUMN allow_direct_messages;

-- Step 9: Drop old indexes that are no longer needed
DROP INDEX IF EXISTS idx_users_pseudonym;
DROP INDEX IF EXISTS idx_users_display_name;
DROP INDEX IF EXISTS idx_users_karma_score;
DROP INDEX IF EXISTS idx_posts_user;
DROP INDEX IF EXISTS idx_comments_user;

-- +migrate Down
-- Rollback migration to restore 1:1 user:pseudonym relationship

-- Step 1: Add back pseudonym-specific columns to users table
ALTER TABLE users ADD COLUMN pseudonym_id VARCHAR(64);
ALTER TABLE users ADD COLUMN display_name VARCHAR(50);
ALTER TABLE users ADD COLUMN karma_score INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN bio TEXT;
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255);
ALTER TABLE users ADD COLUMN website_url VARCHAR(255);
ALTER TABLE users ADD COLUMN show_karma BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN allow_direct_messages BOOLEAN DEFAULT TRUE;

-- Step 2: Restore user data from pseudonyms table (assuming one pseudonym per user)
UPDATE users SET 
    pseudonym_id = p.pseudonym_id,
    display_name = p.display_name,
    karma_score = p.karma_score,
    bio = p.bio,
    avatar_url = p.avatar_url,
    website_url = p.website_url,
    show_karma = p.show_karma,
    allow_direct_messages = p.allow_direct_messages
FROM pseudonyms p
WHERE users.user_id = p.user_id;

-- Step 3: Add constraints back to users table
ALTER TABLE users ALTER COLUMN pseudonym_id SET NOT NULL;
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_pseudonym_id_key UNIQUE (pseudonym_id);

-- Step 4: Restore content tables to use user_id
-- Posts table
ALTER TABLE posts ADD COLUMN user_id BIGINT;
UPDATE posts SET user_id = (
    SELECT user_id FROM pseudonyms WHERE pseudonyms.pseudonym_id = posts.pseudonym_id
);
ALTER TABLE posts ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE posts DROP COLUMN pseudonym_id;
ALTER TABLE posts ADD FOREIGN KEY (user_id) REFERENCES users(user_id);

-- Comments table
ALTER TABLE comments ADD COLUMN user_id BIGINT;
UPDATE comments SET user_id = (
    SELECT user_id FROM pseudonyms WHERE pseudonyms.pseudonym_id = comments.pseudonym_id
);
ALTER TABLE comments ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE comments DROP COLUMN pseudonym_id;
ALTER TABLE comments ADD FOREIGN KEY (user_id) REFERENCES users(user_id);

-- Step 5: Drop foreign key constraints that reference pseudonyms table
-- Drop constraints from votes
ALTER TABLE votes DROP CONSTRAINT IF EXISTS votes_pseudonym_id_fkey;

-- Drop constraints from poll_votes
ALTER TABLE poll_votes DROP CONSTRAINT IF EXISTS poll_votes_pseudonym_id_fkey;

-- Drop constraints from subforum_subscriptions
ALTER TABLE subforum_subscriptions DROP CONSTRAINT IF EXISTS subforum_subscriptions_pseudonym_id_fkey;

-- Drop constraints from user_blocks
ALTER TABLE user_blocks DROP CONSTRAINT IF EXISTS user_blocks_blocker_pseudonym_id_fkey;
ALTER TABLE user_blocks DROP CONSTRAINT IF EXISTS user_blocks_blocked_pseudonym_id_fkey;

-- Drop constraints from direct_messages
ALTER TABLE direct_messages DROP CONSTRAINT IF EXISTS direct_messages_sender_pseudonym_id_fkey;
ALTER TABLE direct_messages DROP CONSTRAINT IF EXISTS direct_messages_recipient_pseudonym_id_fkey;

-- Drop constraints from reports
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_reporter_pseudonym_id_fkey;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_reported_pseudonym_id_fkey;
ALTER TABLE reports DROP CONSTRAINT IF EXISTS reports_resolved_by_pseudonym_id_fkey;

-- Drop constraints from user_bans
ALTER TABLE user_bans DROP CONSTRAINT IF EXISTS user_bans_banned_by_pseudonym_id_fkey;

-- Drop constraints from moderation_actions
ALTER TABLE moderation_actions DROP CONSTRAINT IF EXISTS moderation_actions_moderator_pseudonym_id_fkey;

-- Drop constraints from subforum_moderators
ALTER TABLE subforum_moderators DROP CONSTRAINT IF EXISTS subforum_moderators_pseudonym_id_fkey;

-- Drop constraints from correlation_audit
ALTER TABLE correlation_audit DROP CONSTRAINT IF EXISTS correlation_audit_pseudonym_id_fkey;

-- Drop constraints from posts moderation fields
ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_removed_by_pseudonym_id_fkey;

-- Drop constraints from comments moderation fields
ALTER TABLE comments DROP CONSTRAINT IF EXISTS comments_removed_by_pseudonym_id_fkey;

-- Step 6: Restore foreign key references to users table
-- Update votes table
ALTER TABLE votes ADD FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update poll_votes table
ALTER TABLE poll_votes ADD FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update subforum_subscriptions table
ALTER TABLE subforum_subscriptions ADD FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update user_blocks table
ALTER TABLE user_blocks ADD FOREIGN KEY (blocker_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;
ALTER TABLE user_blocks ADD FOREIGN KEY (blocked_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update direct_messages table
ALTER TABLE direct_messages ADD FOREIGN KEY (sender_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;
ALTER TABLE direct_messages ADD FOREIGN KEY (recipient_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update reports table
ALTER TABLE reports ADD FOREIGN KEY (reporter_pseudonym_id) REFERENCES users(pseudonym_id);
ALTER TABLE reports ADD FOREIGN KEY (reported_pseudonym_id) REFERENCES users(pseudonym_id);
ALTER TABLE reports ADD FOREIGN KEY (resolved_by_pseudonym_id) REFERENCES users(pseudonym_id);

-- Update user_bans table
ALTER TABLE user_bans ADD FOREIGN KEY (banned_by_pseudonym_id) REFERENCES users(pseudonym_id);

-- Update moderation_actions table
ALTER TABLE moderation_actions ADD FOREIGN KEY (moderator_pseudonym_id) REFERENCES users(pseudonym_id);

-- Update subforum_moderators table
ALTER TABLE subforum_moderators ADD FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;

-- Update posts moderation fields
ALTER TABLE posts ADD FOREIGN KEY (removed_by_pseudonym_id) REFERENCES users(pseudonym_id);

-- Update comments moderation fields
ALTER TABLE comments ADD FOREIGN KEY (removed_by_pseudonym_id) REFERENCES users(pseudonym_id);

-- Step 7: Clean up identity_mappings table
ALTER TABLE identity_mappings DROP COLUMN user_id;
ALTER TABLE identity_mappings DROP CONSTRAINT IF EXISTS identity_mappings_user_id_fkey;
ALTER TABLE identity_mappings DROP CONSTRAINT IF EXISTS identity_mappings_pseudonym_id_fkey;
ALTER TABLE identity_mappings ADD FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE;
DROP INDEX IF EXISTS idx_mappings_user;

-- Step 8: Restore old indexes
CREATE INDEX idx_users_pseudonym ON users(pseudonym_id);
CREATE INDEX idx_users_display_name ON users(display_name);
CREATE INDEX idx_users_karma_score ON users(karma_score);
CREATE INDEX idx_posts_user ON posts(user_id);
CREATE INDEX idx_comments_user ON comments(user_id);

-- Step 9: Drop the pseudonyms table
DROP TABLE IF EXISTS pseudonyms;
