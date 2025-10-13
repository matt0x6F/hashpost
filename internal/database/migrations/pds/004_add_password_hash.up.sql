-- Add password hash column to users table
-- This migration adds password hashing support for user authentication

-- Add password hash column
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

-- Add index for password hash lookups (though we'll primarily use handle/email for lookups)
CREATE INDEX idx_users_password_hash ON users(password_hash);

-- Add comment to document the column
COMMENT ON COLUMN users.password_hash IS 'bcrypt hashed password for user authentication';
