-- Drop performance optimization indexes for PDS
-- This migration removes the additional indexes added in 005_performance_indexes.up.sql

-- Drop composite indexes
DROP INDEX IF EXISTS idx_posts_user_subforum_created;
DROP INDEX IF EXISTS idx_posts_subforum_created;
DROP INDEX IF EXISTS idx_comments_post_created;
DROP INDEX IF EXISTS idx_comments_user_created;

-- Drop atproto URI indexes
DROP INDEX IF EXISTS idx_posts_atproto_uri;
DROP INDEX IF EXISTS idx_comments_atproto_uri;

-- Drop vote aggregation indexes
DROP INDEX IF EXISTS idx_votes_post_vote_type;
DROP INDEX IF EXISTS idx_votes_comment_vote_type;

-- Drop subscription lookup indexes
DROP INDEX IF EXISTS idx_subforum_subscriptions_user_created;

-- Drop partial indexes for active records
DROP INDEX IF EXISTS idx_posts_active;
DROP INDEX IF EXISTS idx_comments_active;

-- Drop text search indexes
DROP INDEX IF EXISTS idx_posts_content_search;
DROP INDEX IF EXISTS idx_comments_content_search;
