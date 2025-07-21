# HashPost Database Schema

## Overview

This document defines the complete database schema for HashPost, a Reddit-like social media platform that uses Identity-Based Encryption (IBE) to provide pseudonymous user profiles while maintaining administrative accountability. The schema uses a single-user system with Role-Based Access Control (RBAC) to balance simplicity with security.

## Database Architecture

The HashPost platform uses a **single-database architecture** with **pseudonym-based access control**:

1. **Single Database**: Contains all user data, content, and administrative information
2. **Pseudonym-Based Access**: Different pseudonyms have different capabilities and access levels
3. **Privacy Protection**: Real identities are encrypted and only accessible to users with appropriate roles
4. **Audit Trail**: All administrative activities are logged for compliance and oversight

This approach ensures that regular users cannot correlate pseudonymous profiles while maintaining the ability for authorized users to perform necessary correlation activities based on their role.

## Core User Tables

### `users`
Stores real user identity information (one per actual person).

```sql
CREATE TABLE users (
    user_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    is_suspended BOOLEAN DEFAULT FALSE,
    suspension_reason TEXT,
    suspension_expires_at TIMESTAMP,
    
    -- Basic user roles (platform-wide)
    roles JSONB DEFAULT '["user"]',
    
    -- Administrative fields
    admin_username VARCHAR(100) UNIQUE,
    admin_password_hash VARCHAR(255),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    
    -- Moderation fields
    moderated_subforums JSONB,
    admin_scope VARCHAR(100), -- 'trust_safety', 'legal', 'platform_admin'
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for performance
    INDEX idx_users_email (email),
    INDEX idx_users_admin_username (admin_username),
    INDEX idx_users_roles (roles),
    INDEX idx_users_active (is_active),
    INDEX idx_users_last_active (last_active_at),
    INDEX idx_users_created_at (created_at)
);
```

### `pseudonyms`
Stores pseudonym-specific information (multiple per user) with their own roles and capabilities.

```sql
CREATE TABLE pseudonyms (
    pseudonym_id VARCHAR(64) PRIMARY KEY,
    display_name VARCHAR(50) NOT NULL,
    karma_score INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Pseudonym-specific roles and capabilities
    roles JSONB DEFAULT '["user"]',
    capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]',
    
    -- Profile metadata (optional)
    bio TEXT,
    avatar_url VARCHAR(255),
    website_url VARCHAR(255),
    
    -- Privacy settings
    show_karma BOOLEAN DEFAULT TRUE,
    allow_direct_messages BOOLEAN DEFAULT TRUE,
    
    -- Indexes for performance
    INDEX idx_pseudonyms_active (is_active),
    INDEX idx_pseudonyms_created_at (created_at),
    INDEX idx_pseudonyms_karma_score (karma_score),
    INDEX idx_pseudonyms_last_active (last_active_at),
    UNIQUE CONSTRAINT unique_pseudonym_display_name (display_name)
);
```

### `user_preferences`
Stores user-specific preferences and settings (one per real user).

```sql
CREATE TABLE user_preferences (
    user_id BIGINT PRIMARY KEY,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    theme VARCHAR(20) DEFAULT 'light',
    email_notifications BOOLEAN DEFAULT TRUE,
    push_notifications BOOLEAN DEFAULT TRUE,
    auto_hide_nsfw BOOLEAN DEFAULT TRUE,
    auto_hide_spoilers BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

## Identity Management Tables

### `identity_mappings`
Stores encrypted mappings between real identities and pseudonyms.

```sql
CREATE TABLE identity_mappings (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fingerprint VARCHAR(32) NOT NULL, -- SHA-256 hash of real identity + salt
    user_id BIGINT NOT NULL, -- Real user identity
    pseudonym_id VARCHAR(64) NOT NULL, -- Specific pseudonym
    encrypted_real_identity BYTEA NOT NULL, -- Encrypted email/phone
    encrypted_pseudonym_mapping BYTEA NOT NULL, -- Encrypted mapping data
    key_scope VARCHAR(100) NOT NULL, -- Scope for IBE key usage
    key_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Indexes for administrative lookups
    INDEX idx_mappings_fingerprint (fingerprint),
    INDEX idx_mappings_user (user_id),
    INDEX idx_mappings_pseudonym (pseudonym_id),
    INDEX idx_mappings_key_version (key_version),
    INDEX idx_mappings_created_at (created_at),
    UNIQUE KEY unique_fingerprint_pseudonym (fingerprint, pseudonym_id),
    
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

### `role_keys`
Stores role-based keys for correlation and administrative access.

```sql
CREATE TABLE role_keys (
    key_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(100) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    key_data BYTEA NOT NULL, -- Encrypted key material
    key_version INTEGER NOT NULL DEFAULT 1,
    capabilities JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL,
    
    INDEX idx_role_keys_role (role_name),
    INDEX idx_role_keys_scope (scope),
    INDEX idx_role_keys_expires (expires_at),
    INDEX idx_role_keys_active (is_active),
    
    FOREIGN KEY (created_by) REFERENCES users(user_id)
);
```

## Community Tables

### `subforums`
Stores community subforums (equivalent to Reddit's subreddits).

```sql
CREATE TABLE subforums (
    subforum_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    sidebar_text TEXT,
    rules_text TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
    
    -- Indexes
    INDEX idx_subforums_name (name),
    INDEX idx_subforums_created_at (created_at),
    INDEX idx_subforums_subscriber_count (subscriber_count),
    
    FOREIGN KEY (created_by_user_id) REFERENCES users(user_id)
);
```

### `subforum_subscriptions`
Tracks which pseudonyms are subscribed to which subforums.

```sql
CREATE TABLE subforum_subscriptions (
    subscription_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL,
    subforum_id INTEGER NOT NULL,
    subscribed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_favorite BOOLEAN DEFAULT FALSE,
    
    UNIQUE KEY unique_pseudonym_subforum (pseudonym_id, subforum_id),
    INDEX idx_subscriptions_pseudonym (pseudonym_id),
    INDEX idx_subscriptions_subforum (subforum_id),
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE
);
```

### `subforum_moderators`
Tracks moderator relationships for subforums.

```sql
CREATE TABLE subforum_moderators (
    moderator_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which they moderate
    role VARCHAR(20) NOT NULL DEFAULT 'moderator', -- 'owner', 'moderator', 'junior_moderator'
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    permissions JSONB, -- Store specific permissions as JSONB
    added_by_pseudonym_id VARCHAR(64), -- Pseudonym who added them
    
    UNIQUE KEY unique_subforum_pseudonym (subforum_id, pseudonym_id),
    INDEX idx_moderators_subforum (subforum_id),
    INDEX idx_moderators_pseudonym (pseudonym_id),
    
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (added_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

## Content Tables

### `posts`
Stores all posts made by pseudonyms.

```sql
CREATE TABLE posts (
    post_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym that created the post
    subforum_id INTEGER NOT NULL,
    title VARCHAR(300) NOT NULL,
    content TEXT,
    post_type VARCHAR(20) NOT NULL DEFAULT 'text', -- 'text', 'link', 'image', 'video', 'poll'
    url VARCHAR(2048),
    slug VARCHAR(255), -- URL-friendly slug
    is_self_post BOOLEAN DEFAULT FALSE,
    is_nsfw BOOLEAN DEFAULT FALSE,
    is_spoiler BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    is_stickied BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
    removed_at TIMESTAMP,
    deleted_by_pseudonym_id VARCHAR(64), -- For soft deletes
    
    -- Indexes
    INDEX idx_posts_pseudonym (pseudonym_id),
    INDEX idx_posts_subforum (subforum_id),
    INDEX idx_posts_created_at (created_at),
    INDEX idx_posts_score (score),
    INDEX idx_posts_subforum_created (subforum_id, created_at),
    INDEX idx_posts_subforum_score (subforum_id, score),
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (removed_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (deleted_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

### `comments`
Stores comments on posts.

```sql
CREATE TABLE comments (
    comment_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    post_id BIGINT NOT NULL,
    parent_comment_id BIGINT,
    pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym that created the comment
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    score INTEGER DEFAULT 0,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP,
    edit_reason VARCHAR(100),
    
    -- Moderation fields
    is_removed BOOLEAN DEFAULT FALSE,
    removed_by_user_id BIGINT,
    removed_by_pseudonym_id VARCHAR(64), -- Pseudonym under which removal was performed
    removal_reason VARCHAR(100),
    removed_at TIMESTAMP,
    deleted_by_pseudonym_id VARCHAR(64), -- For soft deletes
    
    -- Indexes
    INDEX idx_comments_post (post_id),
    INDEX idx_comments_parent (parent_comment_id),
    INDEX idx_comments_pseudonym (pseudonym_id),
    INDEX idx_comments_created_at (created_at),
    INDEX idx_comments_score (score),
    INDEX idx_comments_post_score (post_id, score),
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (removed_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (deleted_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

### `votes`
Tracks pseudonym votes on posts and comments.

```sql
CREATE TABLE votes (
    vote_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL,
    content_type VARCHAR(10) NOT NULL, -- 'post' or 'comment'
    content_id BIGINT NOT NULL,
    vote_value INTEGER NOT NULL, -- 1 for upvote, -1 for downvote
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_pseudonym_content_vote (pseudonym_id, content_type, content_id),
    INDEX idx_votes_pseudonym (pseudonym_id),
    INDEX idx_votes_content (content_type, content_id),
    INDEX idx_votes_created_at (created_at),
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

## Media and Attachments

### `media_attachments`
Stores media files attached to posts.

```sql
CREATE TABLE media_attachments (
    attachment_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    post_id BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    width INTEGER,
    height INTEGER,
    duration_seconds INTEGER, -- For videos
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_attachments_post (post_id),
    INDEX idx_attachments_mime_type (mime_type),
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);
```

### `polls`
Stores poll data for poll-type posts.

```sql
CREATE TABLE polls (
    poll_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    post_id BIGINT NOT NULL UNIQUE,
    question TEXT NOT NULL,
    options JSON NOT NULL, -- Array of poll options
    allow_multiple_votes BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);
```

### `poll_votes`
Tracks individual poll votes by pseudonyms.

```sql
CREATE TABLE poll_votes (
    vote_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    poll_id BIGINT NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL,
    selected_options JSON NOT NULL, -- Array of selected option indices
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_pseudonym_poll_vote (poll_id, pseudonym_id),
    INDEX idx_poll_votes_poll (poll_id),
    INDEX idx_poll_votes_pseudonym (pseudonym_id),
    
    FOREIGN KEY (poll_id) REFERENCES polls(poll_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

## User Interaction Tables

### `user_blocks`
Tracks pseudonym blocking relationships. Client-side blocks are always by pseudonym. Backend/admin can correlate and block by user (all personas).

```sql
CREATE TABLE user_blocks (
    block_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    blocker_pseudonym_id VARCHAR(64) NOT NULL,
    blocked_pseudonym_id VARCHAR(64), -- Set for client/user-initiated blocks
    blocked_user_id BIGINT,           -- Set only by backend/admin for 'block all personas'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Enforce that at least one of blocked_pseudonym_id or blocked_user_id is not null, but not both
    CHECK (
        (blocked_pseudonym_id IS NOT NULL AND blocked_user_id IS NULL)
        OR
        (blocked_pseudonym_id IS NULL AND blocked_user_id IS NOT NULL)
    ),

    UNIQUE KEY unique_block_relationship (blocker_pseudonym_id, blocked_pseudonym_id, blocked_user_id),
    INDEX idx_blocks_blocker (blocker_pseudonym_id),
    INDEX idx_blocks_blocked_pseudonym (blocked_pseudonym_id),
    INDEX idx_blocks_blocked_user (blocked_user_id),

    FOREIGN KEY (blocker_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

### `direct_messages`
Stores direct messages between pseudonyms.

```sql
CREATE TABLE direct_messages (
    message_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    sender_pseudonym_id VARCHAR(64) NOT NULL,
    recipient_pseudonym_id VARCHAR(64) NOT NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_messages_sender (sender_pseudonym_id),
    INDEX idx_messages_recipient (recipient_pseudonym_id),
    INDEX idx_messages_created_at (created_at),
    INDEX idx_messages_unread (recipient_pseudonym_id, is_read),
    
    FOREIGN KEY (sender_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

## Moderation Tables

### `reports`
Stores pseudonym reports of content or other pseudonyms.

```sql
CREATE TABLE reports (
    report_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    reporter_pseudonym_id VARCHAR(64) NOT NULL,
    content_type VARCHAR(10) NOT NULL, -- 'post', 'comment', 'user', 'subforum'
    content_id BIGINT,
    reported_pseudonym_id VARCHAR(64),
    report_reason VARCHAR(100) NOT NULL,
    report_details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'investigating', 'resolved', 'dismissed'
    resolved_by_user_id BIGINT, -- Real identity for administrative purposes
    resolved_by_pseudonym_id VARCHAR(64), -- Pseudonym under which report was resolved
    resolution_notes TEXT,
    resolved_at TIMESTAMP,
    
    INDEX idx_reports_reporter (reporter_pseudonym_id),
    INDEX idx_reports_content (content_type, content_id),
    INDEX idx_reports_reported_pseudonym (reported_pseudonym_id),
    INDEX idx_reports_status (status),
    INDEX idx_reports_created_at (created_at),
    
    FOREIGN KEY (reporter_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (reported_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (resolved_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (resolved_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

### `user_bans`
Tracks user bans from subforums.

```sql
CREATE TABLE user_bans (
    ban_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    banned_user_id BIGINT NOT NULL,
    banned_by_user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    banned_by_pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which ban was issued
    ban_reason TEXT NOT NULL,
    is_permanent BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    
    INDEX idx_bans_subforum (subforum_id),
    INDEX idx_bans_banned_user (banned_user_id),
    INDEX idx_bans_banned_by_pseudonym (banned_by_pseudonym_id),
    INDEX idx_bans_expires_at (expires_at),
    INDEX idx_bans_active (is_active),
    
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (banned_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (banned_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (banned_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

### `moderation_actions`
Logs all moderation actions taken by moderators.

```sql
CREATE TABLE moderation_actions (
    action_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    moderator_user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    moderator_pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which action was performed
    subforum_id INTEGER,
    action_type VARCHAR(50) NOT NULL, -- 'remove_post', 'remove_comment', 'ban_user', 'unban_user', etc.
    target_content_type VARCHAR(10), -- 'post', 'comment', 'user'
    target_content_id BIGINT,
    target_user_id BIGINT,
    action_details JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_mod_actions_moderator (moderator_user_id),
    INDEX idx_mod_actions_moderator_pseudonym (moderator_pseudonym_id),
    INDEX idx_mod_actions_subforum (subforum_id),
    INDEX idx_mod_actions_type (action_type),
    INDEX idx_mod_actions_target (target_content_type, target_content_id),
    INDEX idx_mod_actions_created_at (created_at),
    
    FOREIGN KEY (moderator_user_id) REFERENCES users(user_id),
    FOREIGN KEY (moderator_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id) ON DELETE CASCADE,
    FOREIGN KEY (target_user_id) REFERENCES users(user_id)
);
```

## Audit and Compliance Tables

### `correlation_audit`
Logs all correlation activities for compliance and oversight.

```sql
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
    correlation_result JSON,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    legal_basis VARCHAR(100),
    incident_id VARCHAR(100),
    request_source VARCHAR(50), -- 'manual', 'automated', 'api'
    ip_address INET,
    user_agent TEXT,
    
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_pseudonym (pseudonym_id),
    INDEX idx_audit_role (role_used),
    INDEX idx_audit_timestamp (timestamp),
    INDEX idx_audit_incident (incident_id),
    
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

### `key_usage_audit`
Tracks usage of role-based keys for security monitoring.

```sql
CREATE TABLE key_usage_audit (
    usage_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- 'correlation', 'decryption', 'key_rotation'
    target_fingerprint VARCHAR(32),
    target_pseudonym VARCHAR(64),
    success BOOLEAN NOT NULL,
    error_message TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    
    INDEX idx_key_usage_key (key_id),
    INDEX idx_key_usage_user (user_id),
    INDEX idx_key_usage_timestamp (timestamp),
    INDEX idx_key_usage_success (success),
    
    FOREIGN KEY (key_id) REFERENCES role_keys(key_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

### `compliance_reports`
Stores compliance and legal request documentation.

```sql
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
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    notes TEXT,
    
    INDEX idx_compliance_type (report_type),
    INDEX idx_compliance_status (status),
    INDEX idx_compliance_due_date (due_date),
    INDEX idx_compliance_assigned (assigned_user_id),
    
    FOREIGN KEY (assigned_user_id) REFERENCES users(user_id)
);
```

### `compliance_correlations`
Links compliance reports to specific correlation activities.

```sql
CREATE TABLE compliance_correlations (
    correlation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL,
    audit_id UUID NOT NULL,
    correlation_scope TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_compliance_corr_report (report_id),
    INDEX idx_compliance_corr_audit (audit_id),
    
    FOREIGN KEY (report_id) REFERENCES compliance_reports(report_id),
    FOREIGN KEY (audit_id) REFERENCES correlation_audit(audit_id)
);
```

## System Tables

### `system_settings`
Stores global system configuration.

```sql
CREATE TABLE system_settings (
    setting_key VARCHAR(100) PRIMARY KEY,
    setting_value TEXT NOT NULL,
    setting_type VARCHAR(20) NOT NULL DEFAULT 'string', -- 'string', 'integer', 'boolean', 'json'
    description TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by BIGINT,
    
    FOREIGN KEY (updated_by) REFERENCES users(user_id)
);
```

### `api_keys`
Stores API keys for external integrations.

```sql
CREATE TABLE api_keys (
    key_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    key_name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL, -- Associated pseudonym
    permissions JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMP,
    
    INDEX idx_api_keys_hash (key_hash),
    INDEX idx_api_keys_active (is_active),
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

### `system_events`
Logs system-level events for monitoring and debugging.

```sql
CREATE TABLE system_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    event_severity VARCHAR(20) NOT NULL, -- 'info', 'warning', 'error', 'critical'
    event_message TEXT NOT NULL,
    event_data JSON,
    source_component VARCHAR(100),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_events_type (event_type),
    INDEX idx_events_severity (severity),
    INDEX idx_events_timestamp (timestamp),
    INDEX idx_events_component (source_component)
);
```

### `performance_metrics`
Stores performance metrics for system monitoring.

```sql
CREATE TABLE performance_metrics (
    metric_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_unit VARCHAR(20),
    component VARCHAR(100),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_metrics_name (metric_name),
    INDEX idx_metrics_component (component),
    INDEX idx_metrics_timestamp (timestamp)
);
```

## Key Schema Changes

### Recent Migrations Applied

1. **Migration 001**: Initial schema setup
2. **Migration 20250622083429**: Added PostgreSQL functions
3. **Migration 20250622092738**: Multiple pseudonyms support
4. **Migration 20250622183725**: Added pseudonym to API keys
5. **Migration 20250628160133**: Added create subforum capability
6. **Migration 20250628160134**: Removed pseudonym user foreign key
7. **Migration 20250630000212**: Added is_default to pseudonyms
8. **Migration 20250630015219**: Added key_scope to identity mappings
9. **Migration 20250630015701**: Dropped old unique fingerprint pseudonym constraint
10. **Migration 20250630020708**: Fixed identity mappings unique constraint
11. **Migration 20250703000100**: Migrate subforum moderators to IBE
12. **Migration 20250703000200**: Add unique constraint to pseudonym display names
13. **Migration 20250712202336**: Add user role to admin users
14. **Migration 20250712205900**: Add post slug
15. **Migration 20250713000000**: Add user deletion fields
16. **Migration 20250721035243**: Add roles and capabilities to pseudonyms

### Important Schema Notes

- **Users table**: No longer has capabilities field - capabilities moved to pseudonyms
- **Pseudonyms table**: Now has both `roles` and `capabilities` JSONB fields
- **Subforum moderators**: Uses pseudonym_id instead of user_id for moderator identification
- **Posts and comments**: Include slug field for URL-friendly identifiers
- **API keys**: Associated with specific pseudonyms
- **Identity mappings**: Include key_scope field for IBE operations

This schema provides a solid foundation for a Reddit-like platform with IBE-based pseudonymous user profiles while maintaining the administrative capabilities necessary for moderation and compliance through a simplified single-user system with comprehensive RBAC. 