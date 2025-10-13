-- Performance optimization indexes for PDS
-- This migration adds additional indexes for common query patterns

-- Composite indexes for common query patterns
CREATE INDEX idx_posts_user_subforum_created ON posts(user_id, subforum_id, created_at DESC);
CREATE INDEX idx_posts_subforum_created ON posts(subforum_id, created_at DESC);
CREATE INDEX idx_comments_post_created ON comments(post_id, created_at DESC);
CREATE INDEX idx_comments_user_created ON comments(user_id, created_at DESC);

-- Indexes for atproto URI lookups (common in atproto operations)
CREATE INDEX idx_posts_atproto_uri ON posts(atproto_uri) WHERE atproto_uri IS NOT NULL;
CREATE INDEX idx_comments_atproto_uri ON comments(atproto_uri) WHERE atproto_uri IS NOT NULL;

-- Indexes for vote aggregations
CREATE INDEX idx_votes_post_vote_type ON votes(post_id, vote_type) WHERE post_id IS NOT NULL;
CREATE INDEX idx_votes_comment_vote_type ON votes(comment_id, vote_type) WHERE comment_id IS NOT NULL;

-- Indexes for subscription lookups
CREATE INDEX idx_subforum_subscriptions_user_created ON subforum_subscriptions(user_id, created_at DESC);

-- Partial indexes for active records (removed NOW() function as it's not immutable)
-- CREATE INDEX idx_posts_active ON posts(created_at DESC) WHERE created_at > NOW() - INTERVAL '1 year';
-- CREATE INDEX idx_comments_active ON comments(created_at DESC) WHERE created_at > NOW() - INTERVAL '1 year';

-- Indexes for text search (if using full-text search)
CREATE INDEX idx_posts_content_search ON posts USING gin(to_tsvector('english', content));
CREATE INDEX idx_comments_content_search ON comments USING gin(to_tsvector('english', content));
