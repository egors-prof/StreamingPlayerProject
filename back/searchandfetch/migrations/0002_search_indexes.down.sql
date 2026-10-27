DROP INDEX IF EXISTS idx_artists_pseudonym_trgm;
DROP INDEX IF EXISTS idx_albums_title_trgm;
DROP INDEX IF EXISTS idx_songs_title_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
