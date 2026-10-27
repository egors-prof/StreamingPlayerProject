-- Аудиофайл для каждой песни каталога. song_id совпадает с songs.id в БД
-- каталога, но внешнего ключа нет: это другая база (database per service).
CREATE TABLE IF NOT EXISTS tracks
(
    song_id      INTEGER PRIMARY KEY,
    file_name    VARCHAR(255) NOT NULL,
    content_type VARCHAR(64)  NOT NULL DEFAULT 'audio/mpeg',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);
