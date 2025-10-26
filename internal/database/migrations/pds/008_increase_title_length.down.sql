-- Rollback: Restore title column length limit
-- This will truncate any titles longer than 500 characters

ALTER TABLE posts ALTER COLUMN title TYPE CHARACTER VARYING(500);
