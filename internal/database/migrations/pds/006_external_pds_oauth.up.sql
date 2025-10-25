-- External PDS OAuth Support
-- This migration adds support for HashPost to act as an OAuth client to external PDS servers

CREATE TABLE external_pds_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pds_endpoint VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pds_endpoint)
);

-- Index for efficient lookups
CREATE INDEX idx_external_pds_clients_endpoint ON external_pds_clients(pds_endpoint);

-- Add comment
COMMENT ON TABLE external_pds_clients IS 'OAuth client registrations with external PDS servers';
