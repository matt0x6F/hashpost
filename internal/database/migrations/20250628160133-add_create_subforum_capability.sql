-- +migrate Up
-- Add create_subforum capability to regular users
UPDATE role_definitions 
SET capabilities = '["create_content", "vote", "message", "report", "create_subforum"]'::jsonb
WHERE role_name = 'user';

-- +migrate Down
-- Remove create_subforum capability from regular users
UPDATE role_definitions 
SET capabilities = '["create_content", "vote", "message", "report"]'::jsonb
WHERE role_name = 'user';
