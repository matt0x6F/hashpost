-- OAuth Schema for HashPost PDS
-- This migration adds OAuth 2.0 and DPoP support tables

-- OAuth clients table
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    scopes TEXT[] NOT NULL,
    grant_types TEXT[] NOT NULL,
    response_types TEXT[] NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- OAuth authorization codes table
CREATE TABLE oauth_authorization_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    redirect_uri VARCHAR(500) NOT NULL,
    scope TEXT NOT NULL,
    nonce VARCHAR(255),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE
);

-- OAuth access tokens table
CREATE TABLE oauth_access_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    access_token VARCHAR(500) UNIQUE NOT NULL,
    refresh_token VARCHAR(500) UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    scope TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE
);

-- DPoP nonces table
CREATE TABLE dpop_nonces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nonce VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Sessions table for proper session management
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    handle VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_accessed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default OAuth clients
INSERT INTO oauth_clients (client_id, client_name, redirect_uris, scopes, grant_types, response_types) VALUES
('hashpost-web', 'HashPost Web Client', ARRAY['http://localhost:3000/auth/callback'], ARRAY['read', 'write', 'admin'], ARRAY['authorization_code', 'refresh_token'], ARRAY['code']),
('hashpost-mobile', 'HashPost Mobile Client', ARRAY['hashpost://auth/callback'], ARRAY['read', 'write'], ARRAY['authorization_code', 'refresh_token'], ARRAY['code']);

-- Create indexes for performance
CREATE INDEX idx_oauth_authorization_codes_code ON oauth_authorization_codes(code);
CREATE INDEX idx_oauth_authorization_codes_expires_at ON oauth_authorization_codes(expires_at);
CREATE INDEX idx_oauth_access_tokens_access_token ON oauth_access_tokens(access_token);
CREATE INDEX idx_oauth_access_tokens_refresh_token ON oauth_access_tokens(refresh_token);
CREATE INDEX idx_oauth_access_tokens_expires_at ON oauth_access_tokens(expires_at);
CREATE INDEX idx_dpop_nonces_nonce ON dpop_nonces(nonce);
CREATE INDEX idx_dpop_nonces_expires_at ON dpop_nonces(expires_at);
CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX idx_user_sessions_user_did ON user_sessions(user_did);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
