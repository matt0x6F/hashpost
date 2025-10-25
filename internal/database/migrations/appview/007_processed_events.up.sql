-- Create processed_events table for idempotency tracking
CREATE TABLE processed_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) NOT NULL UNIQUE,
    subject VARCHAR(255) NOT NULL,
    sequence BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for fast lookups by event_id
CREATE INDEX idx_processed_events_event_id ON processed_events(event_id);

-- Create index for cleanup of old events
CREATE INDEX idx_processed_events_processed_at ON processed_events(processed_at);

-- Create index for sequence-based queries
CREATE INDEX idx_processed_events_sequence ON processed_events(sequence);
