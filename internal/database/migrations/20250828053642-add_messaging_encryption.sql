
-- +migrate Up
-- Add messaging encryption tables for end-to-end encrypted direct messaging

-- User encryption keys table
CREATE TABLE user_encryption_keys (
    user_id BIGINT PRIMARY KEY,
    encrypted_master_key BYTEA NOT NULL, -- Encrypted with password-derived key
    encrypted_message_keys BYTEA NOT NULL, -- JSON array of encrypted AES keys
    encrypted_signature_key BYTEA NOT NULL, -- Encrypted Ed25519 private key
    public_signature_key BYTEA NOT NULL, -- Ed25519 public key (unencrypted)
    key_fingerprint VARCHAR(64) NOT NULL, -- SHA-256 hash of public key
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_rotated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Conversation encryption keys table
CREATE TABLE conversation_keys (
    conversation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    participant1_user_id BIGINT NOT NULL,
    participant2_user_id BIGINT NOT NULL,
    encrypted_shared_key BYTEA NOT NULL, -- AES-256 key encrypted with both users' public keys
    key_fingerprint VARCHAR(64) NOT NULL, -- Hash of the shared key
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE, -- For forward secrecy
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Ensure consistent ordering of participants
    CONSTRAINT ordered_participants CHECK (
        participant1_user_id < participant2_user_id
    ),
    UNIQUE (participant1_user_id, participant2_user_id),
    
    FOREIGN KEY (participant1_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (participant2_user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Encrypted messages table
CREATE TABLE encrypted_messages (
    message_id BIGINT PRIMARY KEY,
    conversation_id UUID NOT NULL,
    encrypted_content BYTEA NOT NULL, -- AES-256 encrypted message content
    content_hash VARCHAR(64) NOT NULL, -- SHA-256 hash of original content
    iv BYTEA NOT NULL, -- Initialization vector for AES encryption
    signature BYTEA NOT NULL, -- Ed25519 signature for message authenticity
    key_version INTEGER NOT NULL DEFAULT 1, -- Version of encryption key used
    
    FOREIGN KEY (message_id) REFERENCES direct_messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (conversation_id) REFERENCES conversation_keys(conversation_id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_user_encryption_keys_user_id ON user_encryption_keys(user_id);
CREATE INDEX idx_conversation_keys_participants ON conversation_keys(participant1_user_id, participant2_user_id);
CREATE INDEX idx_conversation_keys_expires ON conversation_keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_conversation_keys_active ON conversation_keys(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_encrypted_messages_conversation ON encrypted_messages(conversation_id);
CREATE INDEX idx_encrypted_messages_key_version ON encrypted_messages(key_version);

-- +migrate Down
-- Remove messaging encryption tables

DROP TABLE IF EXISTS encrypted_messages;
DROP TABLE IF EXISTS conversation_keys;
DROP TABLE IF EXISTS user_encryption_keys;
