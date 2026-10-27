-- Поиск идёт по подстроке: ILIKE '%запрос%'. Обычный B-tree индекс при
-- ведущем % не используется, и Postgres перебирал бы всю таблицу.
-- pg_trgm разбивает строки на триграммы, и GIN-индекс по ним ускоряет ILIKE.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_songs_title_trgm ON songs USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_albums_title_trgm ON albums USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_artists_pseudonym_trgm ON artists USING gin (pseudonym gin_trgm_ops);
