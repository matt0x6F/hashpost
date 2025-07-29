-- +migrate Up
-- Add configurable rules system using JSONB for flexibility

-- Add platform rules as JSONB to system_settings
INSERT INTO system_settings (setting_key, setting_value, setting_type, description) VALUES
('platform_rules', '[
  {
    "code": "harassment",
    "name": "No Harassment",
    "description": "Harassment, bullying, or targeted abuse of any kind is not allowed. This includes but is not limited to: threats, intimidation, stalking, and unwanted sexual advances.",
    "category": "safety",
    "severity": "high",
    "active": true
  },
  {
    "code": "hate_speech",
    "name": "No Hate Speech", 
    "description": "Content that promotes hatred, violence, or discrimination against individuals or groups based on race, ethnicity, religion, gender, sexual orientation, disability, or other protected characteristics is prohibited.",
    "category": "safety",
    "severity": "critical",
    "active": true
  },
  {
    "code": "violence",
    "name": "No Violence",
    "description": "Content that promotes, glorifies, or incites violence is not allowed. This includes threats of violence, graphic depictions, or instructions for violent acts.",
    "category": "safety", 
    "severity": "critical",
    "active": true
  },
  {
    "code": "spam",
    "name": "No Spam",
    "description": "Excessive posting of repetitive, unwanted, or promotional content is not allowed. This includes but is not limited to: commercial spam, link farming, and automated posting.",
    "category": "content",
    "severity": "medium", 
    "active": true
  },
  {
    "code": "misinformation",
    "name": "No Misinformation",
    "description": "Deliberately spreading false or misleading information is not allowed. This includes conspiracy theories, medical misinformation, and other harmful falsehoods.",
    "category": "content",
    "severity": "high",
    "active": true
  },
  {
    "code": "privacy_violation",
    "name": "Respect Privacy",
    "description": "Sharing personal information about others without their consent is not allowed. This includes but is not limited to: real names, addresses, phone numbers, and private communications.",
    "category": "safety",
    "severity": "high",
    "active": true
  },
  {
    "code": "copyright",
    "name": "Respect Copyright",
    "description": "Sharing copyrighted content without permission is not allowed. This includes but is not limited to: articles, images, videos, and software.",
    "category": "legal",
    "severity": "medium",
    "active": true
  },
  {
    "code": "illegal_content",
    "name": "No Illegal Content",
    "description": "Content that promotes or facilitates illegal activities is not allowed. This includes but is not limited to: drug sales, weapons, and other illegal goods or services.",
    "category": "legal",
    "severity": "critical",
    "active": true
  }
]', 'json', 'Platform-wide rules for content moderation');

-- Add subforum rules column to subforums table
ALTER TABLE subforums ADD COLUMN subforum_rules JSONB DEFAULT '[]';

-- Enhanced reports table with rule references
ALTER TABLE reports ADD COLUMN rule_code VARCHAR(50);
ALTER TABLE reports ADD COLUMN rule_type VARCHAR(20); -- 'platform' or 'subforum'
ALTER TABLE reports ADD COLUMN forwarded_to_platform BOOLEAN DEFAULT FALSE;
ALTER TABLE reports ADD COLUMN forwarding_notes TEXT;
ALTER TABLE reports ADD COLUMN forwarded_by_user_id BIGINT;
ALTER TABLE reports ADD COLUMN forwarded_at TIMESTAMP WITH TIME ZONE;

-- Add foreign key constraint for forwarded_by
ALTER TABLE reports ADD CONSTRAINT fk_reports_forwarded_by 
    FOREIGN KEY (forwarded_by_user_id) REFERENCES users(user_id);

-- Add indexes for better query performance
CREATE INDEX idx_reports_rule ON reports(rule_code, rule_type);
CREATE INDEX idx_reports_forwarded ON reports(forwarded_to_platform);
CREATE INDEX idx_subforums_subforum_rules ON subforums USING GIN (subforum_rules);

-- +migrate Down
-- Remove configurable rules system

-- Drop indexes
DROP INDEX IF EXISTS idx_reports_forwarded;
DROP INDEX IF EXISTS idx_reports_rule;
DROP INDEX IF EXISTS idx_subforums_subforum_rules;

-- Drop foreign key constraints
ALTER TABLE reports DROP CONSTRAINT IF EXISTS fk_reports_forwarded_by;

-- Remove columns from reports table
ALTER TABLE reports DROP COLUMN IF EXISTS forwarded_at;
ALTER TABLE reports DROP COLUMN IF EXISTS forwarded_by_user_id;
ALTER TABLE reports DROP COLUMN IF EXISTS forwarding_notes;
ALTER TABLE reports DROP COLUMN IF EXISTS forwarded_to_platform;
ALTER TABLE reports DROP COLUMN IF EXISTS rule_type;
ALTER TABLE reports DROP COLUMN IF EXISTS rule_code;

-- Remove subforum rules column from subforums table
ALTER TABLE subforums DROP COLUMN IF EXISTS subforum_rules;

-- Remove platform rules from system_settings
DELETE FROM system_settings WHERE setting_key = 'platform_rules'; 