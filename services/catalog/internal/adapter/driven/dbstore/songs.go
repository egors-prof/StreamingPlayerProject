package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SongRepository struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) *SongRepository { return &SongRepository{db: db} }

type songRow struct {
	ID          int            `db:"id"`
	Title       string         `db:"title"`
	Artist      string         `db:"artist"`
	Album       string         `db:"album"`
	CoverPath   sql.NullString `db:"cover_path"`
	DurationSec int            `db:"duration_sec"`
}

func toDomain(rows []songRow) []domain.Song {
	out := make([]domain.Song, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Song{
			ID: r.ID, Title: r.Title, Artist: r.Artist, Album: r.Album,
			CoverPath: r.CoverPath.String, DurationSec: r.DurationSec,
		})
	}
	return out
}

// Песня всегда принадлежит альбому, а альбом — исполнителю (NOT NULL + FK),
// поэтому обычный JOIN. RIGHT JOIN в старой версии мог вернуть NULL,
// который ломал сканирование в string.
const selectSongs = `
	SELECT s.id, s.title, ar.pseudonym AS artist, al.title AS album,
	       al.cover_path, s.duration_sec
	FROM songs s
	JOIN albums al ON al.id = s.album_id
	JOIN artists ar ON ar.id = al.artist_id`

func (r *SongRepository) List(ctx context.Context, afterID, limit int) ([]domain.Song, error) {
	var rows []songRow
	err := r.db.SelectContext(ctx, &rows, selectSongs+`
		WHERE s.id > $1
		ORDER BY s.id
		LIMIT $2`, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list songs: %w", err)
	}
	return toDomain(rows), nil
}

// Search ищет подстроку в названии песни, исполнителе и альбоме. Сначала
// идут совпадения с начала строки: «neo» поднимет «Neon Rain» выше, чем
// песню, где «neo» встречается в середине.
func (r *SongRepository) Search(ctx context.Context, query string, limit int) ([]domain.Song, error) {
	escaped := escapeLike(query)
	var rows []songRow
	err := r.db.SelectContext(ctx, &rows, selectSongs+`
		WHERE s.title ILIKE $1 OR ar.pseudonym ILIKE $1 OR al.title ILIKE $1
		ORDER BY CASE
		             WHEN s.title ILIKE $2 THEN 0
		             WHEN ar.pseudonym ILIKE $2 THEN 1
		             WHEN al.title ILIKE $2 THEN 2
		             ELSE 3
		         END,
		         s.title, s.id
		LIMIT $3`, "%"+escaped+"%", escaped+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search songs: %w", err)
	}
	return toDomain(rows), nil
}

func (r *SongRepository) GetByIDs(ctx context.Context, ids []int) ([]domain.Song, error) {
	var rows []songRow
	err := r.db.SelectContext(ctx, &rows, selectSongs+`
		WHERE s.id = ANY($1)`, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("get songs by ids: %w", err)
	}
	return toDomain(rows), nil
}

// escapeLike экранирует спецсимволы LIKE. Без этого запрос "%" или "_"
// совпадал бы со всем подряд. Это не SQL-инъекция (значение передаётся
// параметром), но тоже неожиданное поведение.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
