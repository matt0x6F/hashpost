-- RBAC Schema for HashPost AppView
-- This migration adds role-based access control tables

-- Roles table - defines the available roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_platform_role BOOLEAN DEFAULT FALSE, -- true for platform admin, false for subforum-specific
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Permissions table - defines available permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL, -- 'platform', 'subforum', 'post', 'comment'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Role permissions - defines which permissions each role has
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- User roles - assigns roles to users
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL, -- Reference to user's DID
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES appview_subforums(id) ON DELETE CASCADE, -- NULL for platform roles
    granted_by VARCHAR(255) NOT NULL, -- DID of user who granted the role
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ, -- NULL for permanent roles
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_did, role_id, subforum_id)
);

-- Insert default roles
INSERT INTO roles (name, description, is_platform_role) VALUES
('platform_admin', 'Platform administrator with full system access', TRUE),
('subforum_owner', 'Owner of a specific subforum with full control', FALSE),
('subforum_moderator', 'Moderator of a specific subforum with moderation powers', FALSE),
('user', 'Regular user with basic permissions', FALSE);

-- Insert default permissions
INSERT INTO permissions (name, description, resource_type) VALUES
-- Platform permissions
('platform.manage_users', 'Manage all users on the platform', 'platform'),
('platform.manage_subforums', 'Create, modify, and delete subforums', 'platform'),
('platform.view_analytics', 'View platform-wide analytics and reports', 'platform'),
('platform.manage_roles', 'Assign and revoke roles for any user', 'platform'),
('platform.moderate_all', 'Moderate content across all subforums', 'platform'),

-- Subforum permissions
('subforum.manage', 'Full management of a subforum', 'subforum'),
('subforum.moderate', 'Moderate content in a subforum', 'subforum'),
('subforum.create_posts', 'Create posts in a subforum', 'subforum'),
('subforum.create_comments', 'Create comments in a subforum', 'subforum'),
('subforum.view', 'View subforum content', 'subforum'),

-- Post permissions
('post.create', 'Create new posts', 'post'),
('post.edit_own', 'Edit own posts', 'post'),
('post.edit_any', 'Edit any post in subforum', 'post'),
('post.delete_own', 'Delete own posts', 'post'),
('post.delete_any', 'Delete any post in subforum', 'post'),
('post.lock', 'Lock/unlock posts', 'post'),
('post.sticky', 'Sticky/unsticky posts', 'post'),

-- Comment permissions
('comment.create', 'Create new comments', 'comment'),
('comment.edit_own', 'Edit own comments', 'comment'),
('comment.edit_any', 'Edit any comment in subforum', 'comment'),
('comment.delete_own', 'Delete own comments', 'comment'),
('comment.delete_any', 'Delete any comment in subforum', 'comment'),

-- Vote permissions
('vote.create', 'Vote on posts and comments', 'vote');

-- Assign permissions to roles
-- Platform admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'platform_admin';

-- Subforum owner gets subforum management permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'subforum_owner'
AND p.resource_type IN ('subforum', 'post', 'comment', 'vote');

-- Subforum moderator gets moderation permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'subforum_moderator'
AND p.name IN (
    'subforum.moderate', 'subforum.create_posts', 'subforum.create_comments', 'subforum.view',
    'post.edit_any', 'post.delete_any', 'post.lock', 'post.sticky',
    'comment.edit_any', 'comment.delete_any',
    'vote.create'
);

-- Regular user gets basic permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user'
AND p.name IN (
    'subforum.view', 'subforum.create_posts', 'subforum.create_comments',
    'post.create', 'post.edit_own', 'post.delete_own',
    'comment.create', 'comment.edit_own', 'comment.delete_own',
    'vote.create'
);

-- Create indexes for performance
CREATE INDEX idx_user_roles_user_did ON user_roles(user_did);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_subforum_id ON user_roles(subforum_id);
CREATE INDEX idx_user_roles_active ON user_roles(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
