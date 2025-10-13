-- Add atproto_uri field to subforums table
ALTER TABLE subforums ADD COLUMN atproto_uri VARCHAR(500) UNIQUE;
