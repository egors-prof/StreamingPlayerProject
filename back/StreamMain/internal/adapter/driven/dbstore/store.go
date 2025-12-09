package dbstore

import (
	"context"

	"github.com/egors-prof/streaming/internal/domain"
	"github.com/jmoiron/sqlx"
)

type DBStore struct {
	SongStorage *SongStorage
}

func (s *DBStore) GetPathByTitle(ctx context.Context, title string) (*domain.Song, error) {
	dSong, err := s.SongStorage.GetPathByTitle(ctx, title)
	if err != nil {
		return &domain.Song{}, s.SongStorage.translateError(err)
	}
	return dSong, nil
}

func New(db *sqlx.DB) *DBStore {
	return &DBStore{
		SongStorage: NewSongStorage(db),
	}
}
