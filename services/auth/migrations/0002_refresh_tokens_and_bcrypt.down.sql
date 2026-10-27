DROP TABLE IF EXISTS refresh_tokens;
ALTER TABLE users RENAME COLUMN password_hash TO password;
