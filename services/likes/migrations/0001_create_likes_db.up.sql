CREATE TABLE likes (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(45) NOT NULL,
                       song_id INTEGER NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);



-- Индексы для производительности
CREATE INDEX idx_likes_user_id ON likes(username);
CREATE INDEX idx_likes_song_id ON likes(song_id);
CREATE INDEX idx_likes_user_song ON likes(username, song_id);