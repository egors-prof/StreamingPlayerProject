-- Лайки были привязаны к username, который может меняться, а в Go-коде поле
-- было объявлено как int при VARCHAR в БД. Переходим на user_id.
-- Старые строки перенести нельзя: в этой БД нет связи username -> id.
DELETE FROM likes;

DROP INDEX IF EXISTS idx_likes_user_id;
DROP INDEX IF EXISTS idx_likes_user_song;
ALTER TABLE likes DROP COLUMN username;
ALTER TABLE likes ADD COLUMN user_id INTEGER NOT NULL;

-- Один пользователь может лайкнуть песню только один раз. Благодаря этому
-- INSERT ... ON CONFLICT DO NOTHING делает операцию идемпотентной.
ALTER TABLE likes ADD CONSTRAINT likes_user_song_unique UNIQUE (user_id, song_id);

-- Для списка «мои лайки», отсортированного по дате.
CREATE INDEX idx_likes_user_created ON likes (user_id, created_at DESC);
