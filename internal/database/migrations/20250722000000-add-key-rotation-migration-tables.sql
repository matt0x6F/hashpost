-- +migrate Up
-- Add key rotation migration tracking tables for resumable migrations

-- Track overall migration state
CREATE TABLE key_rotation_migrations (
    migration_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(100) NOT NULL,
    old_key_version INTEGER NOT NULL,
    new_key_version INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, in_progress, paused, completed, failed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    paused_at TIMESTAMP WITH TIME ZONE,
    resumed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_records BIGINT,
    processed_records BIGINT DEFAULT 0,
    failed_records BIGINT DEFAULT 0,
    last_processed_id UUID, -- Resume from this point
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_by BIGINT NOT NULL,
    
    UNIQUE(domain, old_key_version, new_key_version),
    FOREIGN KEY (created_by) REFERENCES users(user_id)
);

-- Track individual record migration status
CREATE TABLE migration_progress (
    migration_id UUID NOT NULL,
    mapping_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    PRIMARY KEY (migration_id, mapping_id),
    FOREIGN KEY (migration_id) REFERENCES key_rotation_migrations(migration_id) ON DELETE CASCADE,
    FOREIGN KEY (mapping_id) REFERENCES identity_mappings(mapping_id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_key_rotation_migrations_domain_status ON key_rotation_migrations(domain, status);
CREATE INDEX idx_key_rotation_migrations_status ON key_rotation_migrations(status);
CREATE INDEX idx_key_rotation_migrations_started_at ON key_rotation_migrations(started_at);
CREATE INDEX idx_migration_progress_status ON migration_progress(status);
CREATE INDEX idx_migration_progress_started_at ON migration_progress(started_at);

-- +migrate Down
-- Drop key rotation migration tracking tables

DROP TABLE IF EXISTS migration_progress;
DROP TABLE IF EXISTS key_rotation_migrations; 