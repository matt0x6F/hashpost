-- AppView Sessions table
-- This migration adds the sessions table for tracking user sessions in AppView

-- Sessions table for AppView
CREATE TABLE appview_sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    user_did VARCHAR(255) NOT NULL,
    handle VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_appview_sessions_user_did ON appview_sessions(user_did);
CREATE INDEX idx_appview_sessions_expires_at ON appview_sessions(expires_at);
