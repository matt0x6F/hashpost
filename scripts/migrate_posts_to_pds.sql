-- Migration script to move AppView posts to PDS
-- This script creates posts in the PDS database that were previously only in AppView

-- First, let's see what we're working with
-- AppView posts that need to be migrated to PDS

-- Post 1: Final test with complete database separation
INSERT INTO posts (
    id,
    user_id,
    subforum_id,
    title,
    content,
    atproto_uri,
    created_at,
    updated_at
)
SELECT 
    '8639ce49-8fb7-4fc7-a51f-a8dcb58280fe'::uuid,
    u.id as user_id,
    s.id as subforum_id,
    'Final test with complete database separation',
    'Final test with complete database separation',
    'at://did:plc:hashpost-binding-test/com.hashpost.feed.post/mock-record-id',
    '2025-10-08 11:15:46.539656+00'::timestamptz,
    '2025-10-08 11:15:46.539656+00'::timestamptz
FROM users u, subforums s
WHERE u.did = 'did:plc:hashpost-binding-test'
  AND s.slug = 'general'
ON CONFLICT (id) DO NOTHING;

-- Post 2: Original post content
INSERT INTO posts (
    id,
    user_id,
    subforum_id,
    title,
    content,
    atproto_uri,
    created_at,
    updated_at
)
SELECT 
    '1003ea78-f469-4a8b-b87f-e0eb9a1ee01c'::uuid,
    u.id as user_id,
    s.id as subforum_id,
    'Original post content',
    'Original post content',
    'at://did:plc:hashpost-binding-test/com.hashpost.feed.post/618244d8-0e9e-4e14-95a2-24691d5e97eb',
    '2025-10-08 12:13:36.116147+00'::timestamptz,
    '2025-10-08 12:13:36.116147+00'::timestamptz
FROM users u, subforums s
WHERE u.did = 'did:plc:hashpost-binding-test'
  AND s.slug = 'general'
ON CONFLICT (id) DO NOTHING;

-- Post 3: Creating an atproto forum (the one you're trying to vote on)
INSERT INTO posts (
    id,
    user_id,
    subforum_id,
    title,
    content,
    atproto_uri,
    created_at,
    updated_at
)
SELECT 
    'a0080d2b-41df-441b-a670-a8facd703059'::uuid,
    u.id as user_id,
    s.id as subforum_id,
    'Creating an atproto forum',
    'Hello :wave:

If you''re wondering what HashPost is or strives to be then you''re in the right place! When I first started on HashPost I was focused on building a social network built on a mixture of cryptography and Merkle trees. I was really inspired by people like Maxwell Krohn who has pioneered projects like KeyBase and FOKS. In the time I was building this version of HashPost though our societal circumstances changed drastically. While the threat of unmasking, harassment, and online violence was never greater so too was the chilling of speech in the United States and abroad. Suddenly, data residency really meant something very real and I started to understand the implications of a centrally hosted, privacy oriented social media company.

Too long; didn''t read: There are other ways to stay private on the internet and I think atproto and BlueSky have actually found that middleground through DIDs. After all, a forum without privacy is a dangerous place to be but so too is one where speech is not free.',
    'at://did:plc:hashpost-309ba59a-4d91-41ce-aa07-23b6c49655b4/app.bsky.feed.post/89f61f6f-c6fb-426d-85e0-8da2671ded4f',
    '2025-10-26 01:11:06.872052+00'::timestamptz,
    '2025-10-26 01:11:06.872052+00'::timestamptz
FROM users u, subforums s
WHERE u.did = 'did:plc:hashpost-309ba59a-4d91-41ce-aa07-23b6c49655b4'
  AND s.slug = 'hashpost'
ON CONFLICT (id) DO NOTHING;

-- Verify the migration
SELECT 'Migration completed. Posts in PDS:' as status;
SELECT id, atproto_uri, title, created_at FROM posts ORDER BY created_at;
