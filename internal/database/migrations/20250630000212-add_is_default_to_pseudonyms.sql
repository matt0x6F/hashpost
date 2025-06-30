-- +migrate Up

-- Add is_default column to pseudonyms
ALTER TABLE pseudonyms ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT FALSE;

-- For each user, set is_default = TRUE for their oldest pseudonym
-- (Assume pseudonyms have a created_at timestamp and user_id is available via identity_mappings)
UPDATE pseudonyms SET is_default = TRUE
FROM (
    SELECT im.user_id, p.pseudonym_id
    FROM identity_mappings im
    JOIN pseudonyms p ON p.pseudonym_id = im.pseudonym_id
    WHERE im.is_active = TRUE
    AND p.is_active = TRUE
    AND im.user_id IS NOT NULL
    AND p.created_at IS NOT NULL
    AND im.created_at IS NOT NULL
    AND (
        SELECT MIN(im2.created_at)
        FROM identity_mappings im2
        WHERE im2.user_id = im.user_id
    ) = im.created_at
) oldest
WHERE pseudonyms.pseudonym_id = oldest.pseudonym_id;

-- +migrate Down
