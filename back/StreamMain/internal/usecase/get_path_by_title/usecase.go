package getpathbytitle

import (
	"context"

	"github.com/egors-prof/streaming/internal/config"
	"github.com/egors-prof/streaming/internal/domain"
	"github.com/egors-prof/streaming/internal/port/driven"
)

type UseCase struct {
	cfg         *config.Config
	SongStorage driven.SongStorage
}

func New(cfg *config.Config, songStorage driven.SongStorage) *UseCase {
	return &UseCase{
		cfg:         cfg,
		SongStorage: songStorage,
	}
}

func (u *UseCase) GetPathByTitle(ctx context.Context, title string) (*domain.Song, error) {
	dSong, err := u.SongStorage.GetPathByTitle(ctx, title)
	if err != nil {
		return &domain.Song{}, err
	}
	return dSong, nil
}
