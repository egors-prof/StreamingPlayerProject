-- Демо-каталог. ID песен совпадают с song_id в таблице tracks стрим-сервиса.
INSERT INTO artists (id, pseudonym)
VALUES (1, 'Synthwave Club'),
       (2, 'Aurora Lane'),
       (3, 'Kilo Hertz')
ON CONFLICT DO NOTHING;

INSERT INTO albums (id, artist_id, title, release_year)
VALUES (1, 1, 'Night Drive', 2025),
       (2, 2, 'Glass Horizon', 2024),
       (3, 3, 'Pulse', 2026)
ON CONFLICT DO NOTHING;

INSERT INTO songs (id, album_id, title, duration_sec, track_number)
VALUES (1, 1, 'Neon Rain', 79, 1),
       (2, 1, 'Midnight Arcade', 67, 2),
       (3, 2, 'Glass Horizon', 87, 1),
       (4, 3, 'Pulse Runner', 64, 1),
       (5, 2, 'Low Orbit', 98, 2),
       (6, 3, 'Coastline FM', 73, 2)
ON CONFLICT DO NOTHING;

-- После вставки с явными ID сдвигаем последовательности, иначе следующий
-- INSERT без ID получит id = 1 и упадёт на PRIMARY KEY.
SELECT setval('artists_id_seq', (SELECT max(id) FROM artists));
SELECT setval('albums_id_seq', (SELECT max(id) FROM albums));
SELECT setval('songs_id_seq', (SELECT max(id) FROM songs));
