package dbstore

import (
	"context"

	"github.com/egors-prof/streaming/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SongStorage struct {
	db *sqlx.DB
}

func NewSongStorage(db *sqlx.DB) *SongStorage {
	return &SongStorage{
		db: db,
	}
}

type Song struct {
	ID    int    `db:"id"`
	Title string `db:"title"`
	Path  string `db:"path"`
}

func (rs *Song) FromRepositoryToDomain() *domain.Song {
	return &domain.Song{
		ID:    rs.ID,
		Title: rs.Title,
		Path:  rs.Path,
	}
}

func (s *SongStorage) GetPathByTitle(ctx context.Context, title string) (*domain.Song, error) {
	var rSong Song
	err := s.db.GetContext(ctx, &rSong, `
	select (id,title,path) from streaming;
	`)
	if err != nil {
		return &domain.Song{}, err
	}

	return rSong.FromRepositoryToDomain(), nil
}
