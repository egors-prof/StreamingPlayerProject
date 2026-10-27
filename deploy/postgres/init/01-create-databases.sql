-- Выполняется один раз при первом запуске контейнера Postgres (пустой volume).
-- У каждого сервиса своя база (database per service): сервисы не читают
-- таблицы друг друга и могут менять схему независимо.
CREATE DATABASE auth_db;
CREATE DATABASE likes_db;
CREATE DATABASE catalog_db;
CREATE DATABASE stream_db;
