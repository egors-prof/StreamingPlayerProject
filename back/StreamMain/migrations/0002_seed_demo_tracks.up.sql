-- Демо-треки (сгенерированы scripts/generate_demo_tracks.py, без авторских прав).
-- song_id совпадает с ID песен в seed-миграции каталога (searchandfetch).
INSERT INTO tracks (song_id, file_name, content_type)
VALUES (1, 'demo-01-neon-rain.mp3', 'audio/mpeg'),
       (2, 'demo-02-midnight-arcade.mp3', 'audio/mpeg'),
       (3, 'demo-03-glass-horizon.mp3', 'audio/mpeg'),
       (4, 'demo-04-pulse-runner.mp3', 'audio/mpeg'),
       (5, 'demo-05-low-orbit.mp3', 'audio/mpeg'),
       (6, 'demo-06-coastline-fm.mp3', 'audio/mpeg')
ON CONFLICT (song_id) DO NOTHING;
