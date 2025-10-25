-- Add atproto_uri field to appview_subforums table
ALTER TABLE appview_subforums ADD COLUMN atproto_uri VARCHAR(500);

-- Create index for atproto_uri lookups
CREATE INDEX idx_appview_subforums_atproto_uri ON appview_subforums(atproto_uri);
