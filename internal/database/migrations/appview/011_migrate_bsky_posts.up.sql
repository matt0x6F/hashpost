-- Migrate existing app.bsky.feed.post records to com.hashpost.feed.post in AppView
-- This fixes legacy posts that were created with the wrong collection

UPDATE appview_posts 
SET atproto_uri = REPLACE(atproto_uri, '/app.bsky.feed.post/', '/com.hashpost.feed.post/')
WHERE atproto_uri LIKE '%/app.bsky.feed.post/%';
