-- Migration: add_democratic_moderator_roles
-- +migrate Up
-- Add new role definitions for democratic governance system

INSERT INTO role_definitions (role_name, display_name, description, capabilities, correlation_access, scope, time_window) VALUES
('elected_moderator', 'Elected Moderator', 'Community-elected moderator for democratic subforums', '["moderate_content", "ban_users", "remove_content", "correlate_fingerprints", "manage_moderators", "review_reports", "forward_reports", "manage_subforum_rules", "manage_subforum_settings"]', 'fingerprint', 'subforum_specific', '90_days'),
('appointed_moderator', 'Appointed Moderator', 'Platform-appointed moderator for crisis management', '["moderate_content", "ban_users", "remove_content", "correlate_fingerprints", "manage_moderators", "review_reports", "forward_reports", "manage_subforum_rules", "manage_subforum_settings"]', 'fingerprint', 'subforum_specific', '30_days');

-- +migrate Down
-- Remove democratic moderator role definitions

DELETE FROM role_definitions WHERE role_name IN ('elected_moderator', 'appointed_moderator');