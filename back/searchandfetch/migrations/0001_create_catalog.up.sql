-- Каталог: исполнители → альбомы → песни.
CREATE TABLE IF NOT EXISTS artists
(
    id         SERIAL PRIMARY KEY,
    pseudonym  VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS albums
(
    id           SERIAL PRIMARY KEY,
    artist_id    INTEGER      NOT NULL REFERENCES artists (id) ON DELETE CASCADE,
    title        VARCHAR(200) NOT NULL,
    cover_path   VARCHAR(255),
    release_year SMALLINT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS songs
(
    id           SERIAL PRIMARY KEY,
    album_id     INTEGER      NOT NULL REFERENCES albums (id) ON DELETE CASCADE,
    title        VARCHAR(200) NOT NULL,
    -- Длительность в секундах: число удобно сортировать и складывать,
    -- в отличие от строки, которую раньше приходилось парсить как RFC3339.
    duration_sec INTEGER      NOT NULL CHECK (duration_sec > 0),
    track_number SMALLINT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Postgres не создаёт индексы на внешние ключи сам, а JOIN идёт по ним.
CREATE INDEX IF NOT EXISTS idx_albums_artist_id ON albums (artist_id);
CREATE INDEX IF NOT EXISTS idx_songs_album_id ON songs (album_id);
