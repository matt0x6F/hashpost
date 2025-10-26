-- Rollback: Restore app.bsky.feed.post collection for migrated records in AppView
-- This will revert the URI changes back to app.bsky.feed.post

UPDATE appview_posts 
SET atproto_uri = REPLACE(atproto_uri, '/com.hashpost.feed.post/', '/app.bsky.feed.post/')
WHERE atproto_uri LIKE '%/com.hashpost.feed.post/%' 
AND atproto_uri LIKE '%/89f61f6f-c6fb-426d-85e0-8da2671ded4f%';
