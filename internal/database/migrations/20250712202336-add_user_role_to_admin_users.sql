
-- +migrate Up
-- Add "user" role to existing admin users who only have admin roles
-- This ensures admin users have both basic user capabilities and admin capabilities

-- Update platform_admin users to also have "user" role
UPDATE users 
SET roles = jsonb_build_array('user', 'platform_admin')
WHERE roles @> '["platform_admin"]'::jsonb 
  AND NOT (roles @> '["user"]'::jsonb);

-- Update trust_safety users to also have "user" role  
UPDATE users 
SET roles = jsonb_build_array('user', 'trust_safety')
WHERE roles @> '["trust_safety"]'::jsonb 
  AND NOT (roles @> '["user"]'::jsonb);

-- Update legal_team users to also have "user" role
UPDATE users 
SET roles = jsonb_build_array('user', 'legal_team')
WHERE roles @> '["legal_team"]'::jsonb 
  AND NOT (roles @> '["user"]'::jsonb);

-- Add basic user capabilities to admin users who don't have them
-- Platform admin capabilities should include basic user capabilities
UPDATE users 
SET capabilities = capabilities || '["create_content", "vote", "message", "report", "create_subforum"]'::jsonb
WHERE roles @> '["platform_admin"]'::jsonb 
  AND NOT (capabilities @> '["create_content"]'::jsonb);

-- Trust & Safety capabilities should include basic user capabilities
UPDATE users 
SET capabilities = capabilities || '["create_content", "vote", "message", "report", "create_subforum"]'::jsonb
WHERE roles @> '["trust_safety"]'::jsonb 
  AND NOT (capabilities @> '["create_content"]'::jsonb);

-- Legal team capabilities should include basic user capabilities
UPDATE users 
SET capabilities = capabilities || '["create_content", "vote", "message", "report", "create_subforum"]'::jsonb
WHERE roles @> '["legal_team"]'::jsonb 
  AND NOT (capabilities @> '["create_content"]'::jsonb);

-- +migrate Down
-- Remove "user" role from admin users (revert to admin-only roles)
UPDATE users 
SET roles = jsonb_build_array('platform_admin')
WHERE roles @> '["platform_admin"]'::jsonb 
  AND roles @> '["user"]'::jsonb;

UPDATE users 
SET roles = jsonb_build_array('trust_safety')
WHERE roles @> '["trust_safety"]'::jsonb 
  AND roles @> '["user"]'::jsonb;

UPDATE users 
SET roles = jsonb_build_array('legal_team')
WHERE roles @> '["legal_team"]'::jsonb 
  AND roles @> '["user"]'::jsonb;

-- Remove basic user capabilities from admin users
UPDATE users 
SET capabilities = capabilities - '["create_content", "vote", "message", "report", "create_subforum"]'
WHERE roles @> '["platform_admin"]'::jsonb 
  OR roles @> '["trust_safety"]'::jsonb 
  OR roles @> '["legal_team"]'::jsonb;
