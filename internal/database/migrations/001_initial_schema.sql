-- +migrate Up
-- Initial schema for HashPost Identity-Based Encryption system
-- Complete schema based on database-schema.md specification

-- Core User Tables

-- Users table for pseudonymous user profiles with role-based capabilities
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    pseudonym_id VARCHAR(64) UNIQUE NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    karma_score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    is_suspended BOOLEAN DEFAULT FALSE,
    suspension_reason TEXT,
    suspension_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Profile metadata (optional)
    bio TEXT,
    avatar_url VARCHAR(255),
    website_url VARCHAR(255),
    
    -- Privacy settings
    show_karma BOOLEAN DEFAULT TRUE,
    allow_direct_messages BOOLEAN DEFAULT TRUE,
    
    -- Role-based fields (encrypted in production)
    roles JSONB DEFAULT '["user"]',
    capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]',
    admin_username VARCHAR(100) UNIQUE,
    admin_password_hash VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    
    -- Moderation fields
    moderated_subforums JSONB, -- [{"subforum_id": 1, "role": "moderator"}]
    admin_scope VARCHAR(100), -- 'trust_safety', 'legal', 'platform_admin'
    
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User preferences table
CREATE TABLE user_preferences (
    user_id BIGINT PRIMARY KEY,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    theme VARCHAR(20) DEFAULT 'light',
    email_notifications BOOLEAN DEFAULT TRUE,
    push_notifications BOOLEAN DEFAULT TRUE,
    auto_hide_nsfw BOOLEAN DEFAULT TRUE,
    auto_hide_spoilers BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Identity Management Tables

-- Identity mappings table for encrypted real identity to pseudonym mappings
CREATE TABLE identity_mappings (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fingerprint VARCHAR(32) NOT NULL, -- SHA-256 hash of real identity + salt
    pseudonym_id VARCHAR(64) NOT NULL,
    encrypted_real_identity BYTEA NOT NULL, -- Encrypted email/phone
    encrypted_pseudonym_mapping BYTEA NOT NULL, -- Encrypted mapping data
    key_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Role keys table for role-based correlation access
CREATE TABLE role_keys (
    key_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(100) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    key_data BYTEA NOT NULL, -- Encrypted key material
    key_version INTEGER NOT NULL DEFAULT 1,
    capabilities JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    
    FOREIGN KEY (created_by) REFERENCES users(user_id)
);

-- Community Tables

-- Subforums table for community spaces
CREATE TABLE subforums (
    subforum_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    sidebar_text TEXT,
    rules_text TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by_user_id BIGINT,
    subscriber_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    is_private BOOLEAN DEFAULT FALSE,
    is_restricted BOOLEAN DEFAULT FALSE,
    is_nsfw BOOLEAN DEFAULT FALSE,
    is_quarantined BOOLEAN DEFAULT FALSE,
    
    -- Moderation settings
    allow_images BOOLEAN DEFAULT TRUE,
    allow_videos BOOLEAN DEFAULT TRUE,
    allow_polls BOOLEAN DEFAULT TRUE,
    require_flair BOOLEAN DEFAULT FALSE,
    minimum_account_age_hours INTEGER DEFAULT 0,
    minimum_karma_required INTEGER DEFAULT 0,
    
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (created_by_user_id) REFERENCES users(user_id)
);

-- Subforum subscriptions table
CREATE TABLE subforum_subscriptions (
    subscription_id BIGSERIAL PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL,
    subforum_id INTEGER NOT NULL,
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_favorite BOOLEAN DEFAULT FALSE,
    
    UNIQUE (pseudonym_id, subforum_id),
    
    FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE
);

-- Subforum moderators table
CREATE TABLE subforum_moderators (
    moderator_id BIGSERIAL PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which they moderate
    role VARCHAR(20) NOT NULL DEFAULT 'moderator', -- 'owner', 'moderator', 'junior_moderator'
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    added_by_user_id BIGINT, -- Real identity of who added them
    permissions JSONB, -- Store specific permissions as JSON
    
    UNIQUE (subforum_id, user_id),
    UNIQUE (subforum_id, pseudonym_id),
    
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (added_by_user_id) REFERENCES users(user_id)
);

-- Content Tables

-- Posts table for user content
CREATE TABLE posts (
    post_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subforum_id INTEGER NOT NULL,
    title VARCHAR(300) NOT NULL,
    content TEXT,
    post_type VARCHAR(20) NOT NULL DEFAULT 'text', -- 'text', 'link', 'image', 'video', 'poll'
    url VARCHAR(2048),
    is_self_post BOOLEAN DEFAULT FALSE,
    is_nsfw BOOLEAN DEFAULT FALSE,
    is_spoiler BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    is_stickied BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    score INTEGER DEFAULT 0,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    
    -- Moderation fields
    is_removed BOOLEAN DEFAULT FALSE,
    removed_by_user_id BIGINT, -- Real identity for administrative purposes
    removed_by_pseudonym_id VARCHAR(64), -- Pseudonym under which removal was performed
    removal_reason VARCHAR(100),
    removed_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (removed_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_pseudonym_id) REFERENCES users(pseudonym_id)
);

-- Comments table
CREATE TABLE comments (
    comment_id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    parent_comment_id BIGINT,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    score INTEGER DEFAULT 0,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    edit_reason VARCHAR(100),
    
    -- Moderation fields
    is_removed BOOLEAN DEFAULT FALSE,
    removed_by_user_id BIGINT,
    removed_by_pseudonym_id VARCHAR(64), -- Pseudonym under which removal was performed
    removal_reason VARCHAR(100),
    removed_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_pseudonym_id) REFERENCES users(pseudonym_id)
);

-- Votes table
CREATE TABLE votes (
    vote_id BIGSERIAL PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL,
    content_type VARCHAR(10) NOT NULL, -- 'post' or 'comment'
    content_id BIGINT NOT NULL,
    vote_value INTEGER NOT NULL, -- 1 for upvote, -1 for downvote
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE (pseudonym_id, content_type, content_id),
    
    FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE
);

-- Media and Attachments

-- Media attachments table
CREATE TABLE media_attachments (
    attachment_id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    width INTEGER,
    height INTEGER,
    duration_seconds INTEGER, -- For videos
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);

-- Polls table
CREATE TABLE polls (
    poll_id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL UNIQUE,
    question TEXT NOT NULL,
    options JSONB NOT NULL, -- Array of poll options
    allow_multiple_votes BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);

-- Poll votes table
CREATE TABLE poll_votes (
    vote_id BIGSERIAL PRIMARY KEY,
    poll_id BIGINT NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL,
    selected_options JSONB NOT NULL, -- Array of selected option indices
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE (poll_id, pseudonym_id),
    
    FOREIGN KEY (poll_id) REFERENCES polls(poll_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE
);

-- User Interaction Tables

-- User blocks table
CREATE TABLE user_blocks (
    block_id BIGSERIAL PRIMARY KEY,
    blocker_pseudonym_id VARCHAR(64) NOT NULL,
    blocked_pseudonym_id VARCHAR(64), -- Set for client/user-initiated blocks
    blocked_user_id BIGINT,           -- Set only by backend/admin for 'block all personas'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Enforce that at least one of blocked_pseudonym_id or blocked_user_id is not null, but not both
    CHECK (
        (blocked_pseudonym_id IS NOT NULL AND blocked_user_id IS NULL)
        OR
        (blocked_pseudonym_id IS NULL AND blocked_user_id IS NOT NULL)
    ),

    UNIQUE (blocker_pseudonym_id, blocked_pseudonym_id, blocked_user_id),
    
    FOREIGN KEY (blocker_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Direct messages table
CREATE TABLE direct_messages (
    message_id BIGSERIAL PRIMARY KEY,
    sender_pseudonym_id VARCHAR(64) NOT NULL,
    recipient_pseudonym_id VARCHAR(64) NOT NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (sender_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_pseudonym_id) REFERENCES users(pseudonym_id) ON DELETE CASCADE
);

-- Moderation Tables

-- Reports table
CREATE TABLE reports (
    report_id BIGSERIAL PRIMARY KEY,
    reporter_pseudonym_id VARCHAR(64) NOT NULL,
    content_type VARCHAR(10) NOT NULL, -- 'post', 'comment', 'user', 'subforum'
    content_id BIGINT,
    reported_pseudonym_id VARCHAR(64),
    report_reason VARCHAR(100) NOT NULL,
    report_details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'investigating', 'resolved', 'dismissed'
    resolved_by_user_id BIGINT, -- Real identity for administrative purposes
    resolved_by_pseudonym_id VARCHAR(64), -- Pseudonym under which report was resolved
    resolution_notes TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (reporter_pseudonym_id) REFERENCES users(pseudonym_id),
    FOREIGN KEY (reported_pseudonym_id) REFERENCES users(pseudonym_id),
    FOREIGN KEY (resolved_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (resolved_by_pseudonym_id) REFERENCES users(pseudonym_id)
);

-- User bans table
CREATE TABLE user_bans (
    ban_id BIGSERIAL PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    banned_user_id BIGINT NOT NULL,
    banned_by_user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    banned_by_pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which ban was issued
    ban_reason TEXT NOT NULL,
    is_permanent BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (banned_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (banned_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (banned_by_pseudonym_id) REFERENCES users(pseudonym_id)
);

-- Moderation actions table
CREATE TABLE moderation_actions (
    action_id BIGSERIAL PRIMARY KEY,
    moderator_user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    moderator_pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which action was performed
    subforum_id INTEGER,
    action_type VARCHAR(50) NOT NULL, -- 'remove_post', 'remove_comment', 'ban_user', 'unban_user', etc.
    target_content_type VARCHAR(10), -- 'post', 'comment', 'user'
    target_content_id BIGINT,
    target_user_id BIGINT,
    action_details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (moderator_user_id) REFERENCES users(user_id),
    FOREIGN KEY (moderator_pseudonym_id) REFERENCES users(pseudonym_id),
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (target_user_id) REFERENCES users(user_id)
);

-- Audit and Compliance Tables

-- Correlation audit table for logging all correlation activities
CREATE TABLE correlation_audit (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL,
    admin_username VARCHAR(100) NOT NULL,
    role_used VARCHAR(50) NOT NULL, -- 'moderator', 'admin'
    requested_pseudonym VARCHAR(64) NOT NULL,
    requested_fingerprint VARCHAR(32),
    justification TEXT NOT NULL,
    correlation_type VARCHAR(20) NOT NULL, -- 'fingerprint', 'identity'
    correlation_result JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    legal_basis VARCHAR(100),
    incident_id VARCHAR(100),
    request_source VARCHAR(50), -- 'manual', 'automated', 'api'
    ip_address INET,
    user_agent TEXT,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Key usage audit table
CREATE TABLE key_usage_audit (
    usage_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- 'correlation', 'decryption', 'key_rotation'
    target_fingerprint VARCHAR(32),
    target_pseudonym VARCHAR(64),
    success BOOLEAN NOT NULL,
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    
    FOREIGN KEY (key_id) REFERENCES role_keys(key_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Compliance reports table
CREATE TABLE compliance_reports (
    report_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_type VARCHAR(50) NOT NULL, -- 'court_order', 'subpoena', 'law_enforcement', 'internal_audit'
    requesting_authority VARCHAR(255),
    request_id VARCHAR(100),
    request_date DATE NOT NULL,
    due_date DATE,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'rejected'
    scope_description TEXT NOT NULL,
    legal_basis TEXT,
    assigned_user_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    
    FOREIGN KEY (assigned_user_id) REFERENCES users(user_id)
);

-- Compliance correlations table
CREATE TABLE compliance_correlations (
    correlation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL,
    audit_id UUID NOT NULL,
    correlation_scope TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (report_id) REFERENCES compliance_reports(report_id),
    FOREIGN KEY (audit_id) REFERENCES correlation_audit(audit_id)
);

-- System Tables

-- System settings table
CREATE TABLE system_settings (
    setting_key VARCHAR(100) PRIMARY KEY,
    setting_value TEXT NOT NULL,
    setting_type VARCHAR(20) NOT NULL DEFAULT 'string', -- 'string', 'integer', 'boolean', 'json'
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by BIGINT,
    
    FOREIGN KEY (updated_by) REFERENCES users(user_id)
);

-- API keys table
CREATE TABLE api_keys (
    key_id BIGSERIAL PRIMARY KEY,
    key_name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    permissions JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- System events table
CREATE TABLE system_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    event_severity VARCHAR(20) NOT NULL, -- 'info', 'warning', 'error', 'critical'
    event_message TEXT NOT NULL,
    event_data JSONB,
    source_component VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Performance metrics table
CREATE TABLE performance_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(20),
    component VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Role-Based Access Control

-- Role definitions table
CREATE TABLE role_definitions (
    role_id SERIAL PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    capabilities JSONB NOT NULL,
    correlation_access VARCHAR(20), -- 'none', 'fingerprint', 'identity'
    scope VARCHAR(100), -- 'none', 'subforum_specific', 'platform_wide'
    time_window VARCHAR(20), -- 'none', '30_days', '90_days', 'unlimited'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO role_definitions (role_name, display_name, description, capabilities, correlation_access, scope, time_window) VALUES
('user', 'Regular User', 'Standard platform user', '["create_content", "vote", "message", "report"]', 'none', 'none', 'none'),
('moderator', 'Subforum Moderator', 'Moderator for specific subforums', '["moderate_content", "ban_users", "remove_content", "correlate_fingerprints"]', 'fingerprint', 'subforum_specific', '30_days'),
('subforum_owner', 'Subforum Owner', 'Owner of a subforum', '["moderate_content", "ban_users", "remove_content", "correlate_fingerprints", "manage_moderators"]', 'fingerprint', 'subforum_specific', '90_days'),
('trust_safety', 'Trust & Safety', 'Platform-wide safety and harassment investigation', '["correlate_identities", "cross_platform_access", "system_moderation"]', 'identity', 'platform_wide', 'unlimited'),
('legal_team', 'Legal Team', 'Legal compliance and court order handling', '["correlate_identities", "legal_compliance", "court_orders"]', 'identity', 'platform_wide', 'unlimited'),
('platform_admin', 'Platform Administrator', 'Full system administration', '["system_admin", "user_management", "correlate_identities"]', 'identity', 'platform_wide', 'unlimited');

-- Indexes for better query performance

-- User indexes
CREATE INDEX idx_users_pseudonym ON users(pseudonym_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_display_name ON users(display_name);
CREATE INDEX idx_users_karma_score ON users(karma_score);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_last_active ON users(last_active_at);
CREATE INDEX idx_users_admin_username ON users(admin_username);
CREATE INDEX idx_users_roles ON users USING GIN(roles);
CREATE INDEX idx_users_active ON users(is_active);

-- Identity mapping indexes
CREATE INDEX idx_mappings_fingerprint ON identity_mappings(fingerprint);
CREATE INDEX idx_mappings_pseudonym ON identity_mappings(pseudonym_id);
CREATE INDEX idx_mappings_key_version ON identity_mappings(key_version);
CREATE INDEX idx_mappings_created_at ON identity_mappings(created_at);
CREATE UNIQUE INDEX unique_fingerprint_pseudonym ON identity_mappings(fingerprint, pseudonym_id);

-- Role key indexes
CREATE INDEX idx_role_keys_role ON role_keys(role_name);
CREATE INDEX idx_role_keys_scope ON role_keys(scope);
CREATE INDEX idx_role_keys_expires ON role_keys(expires_at);
CREATE INDEX idx_role_keys_active ON role_keys(is_active);

-- Subforum indexes
CREATE INDEX idx_subforums_name ON subforums(name);
CREATE INDEX idx_subforums_created_at ON subforums(created_at);
CREATE INDEX idx_subforums_subscriber_count ON subforums(subscriber_count);

-- Subscription indexes
CREATE INDEX idx_subscriptions_pseudonym ON subforum_subscriptions(pseudonym_id);
CREATE INDEX idx_subscriptions_subforum ON subforum_subscriptions(subforum_id);

-- Moderator indexes
CREATE INDEX idx_moderators_subforum ON subforum_moderators(subforum_id);
CREATE INDEX idx_moderators_user ON subforum_moderators(user_id);
CREATE INDEX idx_moderators_pseudonym ON subforum_moderators(pseudonym_id);

-- Post indexes
CREATE INDEX idx_posts_user ON posts(user_id);
CREATE INDEX idx_posts_subforum ON posts(subforum_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_posts_score ON posts(score);
CREATE INDEX idx_posts_subforum_created ON posts(subforum_id, created_at);
CREATE INDEX idx_posts_subforum_score ON posts(subforum_id, score);

-- Comment indexes
CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_parent ON comments(parent_comment_id);
CREATE INDEX idx_comments_user ON comments(user_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_comments_score ON comments(score);
CREATE INDEX idx_comments_post_score ON comments(post_id, score);

-- Vote indexes
CREATE INDEX idx_votes_pseudonym ON votes(pseudonym_id);
CREATE INDEX idx_votes_content ON votes(content_type, content_id);
CREATE INDEX idx_votes_created_at ON votes(created_at);

-- Media attachment indexes
CREATE INDEX idx_attachments_post ON media_attachments(post_id);
CREATE INDEX idx_attachments_mime_type ON media_attachments(mime_type);

-- Poll vote indexes
CREATE INDEX idx_poll_votes_poll ON poll_votes(poll_id);
CREATE INDEX idx_poll_votes_pseudonym ON poll_votes(pseudonym_id);

-- User block indexes
CREATE INDEX idx_blocks_blocker ON user_blocks(blocker_pseudonym_id);
CREATE INDEX idx_blocks_blocked_pseudonym ON user_blocks(blocked_pseudonym_id);
CREATE INDEX idx_blocks_blocked_user ON user_blocks(blocked_user_id);

-- Message indexes
CREATE INDEX idx_messages_sender ON direct_messages(sender_pseudonym_id);
CREATE INDEX idx_messages_recipient ON direct_messages(recipient_pseudonym_id);
CREATE INDEX idx_messages_created_at ON direct_messages(created_at);
CREATE INDEX idx_messages_unread ON direct_messages(recipient_pseudonym_id, is_read);

-- Report indexes
CREATE INDEX idx_reports_reporter ON reports(reporter_pseudonym_id);
CREATE INDEX idx_reports_content ON reports(content_type, content_id);
CREATE INDEX idx_reports_reported_pseudonym ON reports(reported_pseudonym_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_created_at ON reports(created_at);

-- Ban indexes
CREATE INDEX idx_bans_subforum ON user_bans(subforum_id);
CREATE INDEX idx_bans_banned_user ON user_bans(banned_user_id);
CREATE INDEX idx_bans_banned_by_pseudonym ON user_bans(banned_by_pseudonym_id);
CREATE INDEX idx_bans_expires_at ON user_bans(expires_at);
CREATE INDEX idx_bans_active ON user_bans(is_active);

-- Moderation action indexes
CREATE INDEX idx_mod_actions_moderator ON moderation_actions(moderator_user_id);
CREATE INDEX idx_mod_actions_moderator_pseudonym ON moderation_actions(moderator_pseudonym_id);
CREATE INDEX idx_mod_actions_subforum ON moderation_actions(subforum_id);
CREATE INDEX idx_mod_actions_type ON moderation_actions(action_type);
CREATE INDEX idx_mod_actions_target ON moderation_actions(target_content_type, target_content_id);
CREATE INDEX idx_mod_actions_created_at ON moderation_actions(created_at);

-- Audit indexes
CREATE INDEX idx_audit_user ON correlation_audit(user_id);
CREATE INDEX idx_audit_pseudonym ON correlation_audit(pseudonym_id);
CREATE INDEX idx_audit_role ON correlation_audit(role_used);
CREATE INDEX idx_audit_timestamp ON correlation_audit(timestamp);
CREATE INDEX idx_audit_incident ON correlation_audit(incident_id);

-- Key usage audit indexes
CREATE INDEX idx_key_usage_key ON key_usage_audit(key_id);
CREATE INDEX idx_key_usage_user ON key_usage_audit(user_id);
CREATE INDEX idx_key_usage_timestamp ON key_usage_audit(timestamp);
CREATE INDEX idx_key_usage_success ON key_usage_audit(success);

-- Compliance indexes
CREATE INDEX idx_compliance_type ON compliance_reports(report_type);
CREATE INDEX idx_compliance_status ON compliance_reports(status);
CREATE INDEX idx_compliance_due_date ON compliance_reports(due_date);
CREATE INDEX idx_compliance_assigned ON compliance_reports(assigned_user_id);

-- Compliance correlation indexes
CREATE INDEX idx_compliance_corr_report ON compliance_correlations(report_id);
CREATE INDEX idx_compliance_corr_audit ON compliance_correlations(audit_id);

-- API key indexes
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);

-- System event indexes
CREATE INDEX idx_events_type ON system_events(event_type);
CREATE INDEX idx_events_severity ON system_events(event_severity);
CREATE INDEX idx_events_timestamp ON system_events(timestamp);
CREATE INDEX idx_events_component ON system_events(source_component);

-- Performance metric indexes
CREATE INDEX idx_metrics_name ON performance_metrics(metric_name);
CREATE INDEX idx_metrics_component ON performance_metrics(component);
CREATE INDEX idx_metrics_timestamp ON performance_metrics(timestamp);

-- +migrate Down
-- Drop all functions first
DROP FUNCTION IF EXISTS has_capability(BIGINT, VARCHAR(50));
DROP FUNCTION IF EXISTS can_correlate(BIGINT, VARCHAR(20), VARCHAR(100));
DROP FUNCTION IF EXISTS require_mfa_for_action(BIGINT, VARCHAR(50));

-- Drop all tables in reverse dependency order
DROP TABLE IF EXISTS performance_metrics;
DROP TABLE IF EXISTS system_events;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS compliance_correlations;
DROP TABLE IF EXISTS compliance_reports;
DROP TABLE IF EXISTS key_usage_audit;
DROP TABLE IF EXISTS correlation_audit;
DROP TABLE IF EXISTS moderation_actions;
DROP TABLE IF EXISTS user_bans;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS direct_messages;
DROP TABLE IF EXISTS user_blocks;
DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS polls;
DROP TABLE IF EXISTS media_attachments;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS subforum_moderators;
DROP TABLE IF EXISTS subforum_subscriptions;
DROP TABLE IF EXISTS subforums;
DROP TABLE IF EXISTS role_keys;
DROP TABLE IF EXISTS identity_mappings;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_definitions; 