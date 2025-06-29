-- +migrate Up
-- Add permission and correlation checking functions

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION has_capability(
    p_user_id BIGINT,
    p_capability VARCHAR(50)
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM users 
        WHERE user_id = p_user_id 
        AND is_active = TRUE
        AND capabilities @> jsonb_build_array(p_capability)
    );
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION can_correlate(
    p_user_id BIGINT,
    p_correlation_type VARCHAR(20), -- 'fingerprint' or 'identity'
    p_scope VARCHAR(100) DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM users u
        JOIN role_definitions rd ON u.roles @> jsonb_build_array(rd.role_name)
        WHERE u.user_id = p_user_id 
        AND u.is_active = TRUE
        AND rd.correlation_access = p_correlation_type
        AND (p_scope IS NULL OR rd.scope = p_scope)
    );
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION require_mfa_for_action(
    p_user_id BIGINT,
    p_action VARCHAR(50)
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM users 
        WHERE user_id = p_user_id 
        AND mfa_enabled = TRUE
        AND (
            p_action IN ('correlate_identities', 'system_admin', 'legal_compliance')
            OR (p_action = 'correlate_fingerprints' AND admin_scope IS NOT NULL)
        )
    );
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION IF EXISTS has_capability(BIGINT, VARCHAR);
DROP FUNCTION IF EXISTS can_correlate(BIGINT, VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS require_mfa_for_action(BIGINT, VARCHAR);
