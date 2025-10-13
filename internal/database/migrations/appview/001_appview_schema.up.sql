-- AppView Database Schema
-- This migration creates the denormalized schema for the AppView component
-- Optimized for read-heavy operations and user-facing queries

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table (denormalized for AppView)
CREATE TABLE appview_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    reputation INTEGER DEFAULT 0
);

-- Subforums table (denormalized for AppView)
CREATE TABLE appview_subforums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_by_did VARCHAR(255) NOT NULL,
    created_by_handle VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    subscriber_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0
);

-- Posts table (denormalized for AppView)
CREATE TABLE appview_posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    subforum_slug VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0 -- calculated: upvotes - downvotes
);

-- Comments table (denormalized for AppView)
CREATE TABLE appview_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    post_id UUID REFERENCES appview_posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES appview_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0
);

-- Subscriptions table (denormalized for AppView)
CREATE TABLE appview_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL,
    user_handle VARCHAR(255) NOT NULL,
    subforum_slug VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_did, subforum_slug)
);

-- Votes table (denormalized for AppView)
CREATE TABLE appview_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL,
    post_id UUID REFERENCES appview_posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES appview_comments(id) ON DELETE CASCADE,
    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('up', 'down')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_did, post_id),
    UNIQUE(user_did, comment_id),
    CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR 
        (post_id IS NULL AND comment_id IS NOT NULL)
    )
);

-- Indexes for performance
CREATE INDEX idx_appview_users_did ON appview_users(did);
CREATE INDEX idx_appview_users_handle ON appview_users(handle);
CREATE INDEX idx_appview_subforums_slug ON appview_subforums(slug);
CREATE INDEX idx_appview_posts_subforum_slug ON appview_posts(subforum_slug);
CREATE INDEX idx_appview_posts_author_did ON appview_posts(author_did);
CREATE INDEX idx_appview_posts_created_at ON appview_posts(created_at DESC);
CREATE INDEX idx_appview_posts_score ON appview_posts(score DESC);
CREATE INDEX idx_appview_comments_post_id ON appview_comments(post_id);
CREATE INDEX idx_appview_comments_author_did ON appview_comments(author_did);
CREATE INDEX idx_appview_comments_created_at ON appview_comments(created_at DESC);
CREATE INDEX idx_appview_subscriptions_user_did ON appview_subscriptions(user_did);
CREATE INDEX idx_appview_subscriptions_subforum_slug ON appview_subscriptions(subforum_slug);
CREATE INDEX idx_appview_votes_post_id ON appview_votes(post_id);
CREATE INDEX idx_appview_votes_comment_id ON appview_votes(comment_id);
CREATE INDEX idx_appview_votes_user_did ON appview_votes(user_did);
