-- Performance optimization indexes for AppView
-- This migration adds additional indexes for common query patterns in the AppView database

-- Composite indexes for common query patterns
CREATE INDEX idx_appview_posts_author_created ON appview_posts(author_did, created_at DESC);
CREATE INDEX idx_appview_posts_subforum_author_created ON appview_posts(subforum_slug, author_did, created_at DESC);
CREATE INDEX idx_appview_posts_subforum_score ON appview_posts(subforum_slug, score DESC, created_at DESC);
CREATE INDEX idx_appview_comments_post_created ON appview_comments(post_id, created_at DESC);
CREATE INDEX idx_appview_comments_author_created ON appview_comments(author_did, created_at DESC);

-- Indexes for atproto URI lookups (common in atproto operations)
CREATE INDEX idx_appview_posts_atproto_uri ON appview_posts(atproto_uri);
CREATE INDEX idx_appview_comments_atproto_uri ON appview_comments(atproto_uri);

-- Indexes for vote aggregations
CREATE INDEX idx_appview_votes_post_vote_type ON appview_votes(post_id, vote_type) WHERE post_id IS NOT NULL;
CREATE INDEX idx_appview_votes_comment_vote_type ON appview_votes(comment_id, vote_type) WHERE comment_id IS NOT NULL;

-- Indexes for subscription lookups
CREATE INDEX idx_appview_subscriptions_user_created ON appview_subscriptions(user_did, created_at DESC);
CREATE INDEX idx_appview_subscriptions_subforum_created ON appview_subscriptions(subforum_slug, created_at DESC);

-- Partial indexes for active records (removed NOW() function as it's not immutable)
-- CREATE INDEX idx_appview_posts_active ON appview_posts(created_at DESC) WHERE created_at > NOW() - INTERVAL '1 year';
-- CREATE INDEX idx_appview_comments_active ON appview_comments(created_at DESC) WHERE created_at > NOW() - INTERVAL '1 year';

-- Indexes for text search (if using full-text search)
CREATE INDEX idx_appview_posts_content_search ON appview_posts USING gin(to_tsvector('english', content));
CREATE INDEX idx_appview_comments_content_search ON appview_comments USING gin(to_tsvector('english', content));

-- Indexes for user stats queries
CREATE INDEX idx_appview_users_reputation ON appview_users(reputation DESC);
CREATE INDEX idx_appview_users_post_count ON appview_users(post_count DESC);

-- Indexes for subforum stats queries
CREATE INDEX idx_appview_subforums_subscriber_count ON appview_subforums(subscriber_count DESC);
CREATE INDEX idx_appview_subforums_post_count ON appview_subforums(post_count DESC);
