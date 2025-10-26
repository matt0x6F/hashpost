-- Migrate existing app.bsky.feed.post records to com.hashpost.feed.post
-- This fixes legacy posts that were created with the wrong collection

UPDATE posts 
SET atproto_uri = REPLACE(atproto_uri, '/app.bsky.feed.post/', '/com.hashpost.feed.post/')
WHERE atproto_uri LIKE '%/app.bsky.feed.post/%';
