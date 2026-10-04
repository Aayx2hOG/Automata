ALTER TABLE users ADD COLUMN username TEXT;

UPDATE users
SET username = 'user_' || replace(id::text, '-', '')
WHERE username IS NULL;

ALTER TABLE users ALTER COLUMN username SET NOT NULL;
CREATE UNIQUE INDEX idx_users_username ON users(username);
