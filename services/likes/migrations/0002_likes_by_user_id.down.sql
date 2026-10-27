DROP INDEX IF EXISTS idx_likes_user_created;
ALTER TABLE likes DROP CONSTRAINT IF EXISTS likes_user_song_unique;
DELETE FROM likes;
ALTER TABLE likes DROP COLUMN user_id;
ALTER TABLE likes ADD COLUMN username VARCHAR(45) NOT NULL;
CREATE INDEX idx_likes_user_id ON likes (username);
CREATE INDEX idx_likes_user_song ON likes (username, song_id);
