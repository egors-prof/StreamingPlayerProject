package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/domain"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/errs"
	"github.com/jmoiron/sqlx"
)

type TrackStorage struct {
	db *sqlx.DB
}

func NewTrackStorage(db *sqlx.DB) *TrackStorage { return &TrackStorage{db: db} }

type trackRow struct {
	SongID      int    `db:"song_id"`
	FileName    string `db:"file_name"`
	ContentType string `db:"content_type"`
}

func (s *TrackStorage) GetTrack(ctx context.Context, songID int) (domain.Track, error) {
	var row trackRow
	err := s.db.GetContext(ctx, &row,
		`SELECT song_id, file_name, content_type FROM tracks WHERE song_id = $1`, songID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Track{}, errs.ErrNotFound
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("select track: %w", err)
	}
	return domain.Track{SongID: row.SongID, FileName: row.FileName, ContentType: row.ContentType}, nil
}
