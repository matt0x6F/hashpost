-- Increase title column length to support long-form content
-- HashPost is designed for long-form content, so titles can be much longer than 500 characters

ALTER TABLE posts ALTER COLUMN title TYPE TEXT;
