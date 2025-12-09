package likeSetter

import (
	"context"

	"github.com/egors-prof/likes_service/internal/config"
	"github.com/egors-prof/likes_service/internal/port/driven"
)

type UseCase struct {
	cfg         *config.Config
	LikeStorage driven.LikeStorage
}

func New(cfg *config.Config, ls driven.LikeStorage) *UseCase {
	return &UseCase{
		cfg:         cfg,
		LikeStorage: ls,
	}
}

func (u *UseCase) SetLike(ctx context.Context, songID int, username string) error {
	err := u.LikeStorage.SetLike(ctx, songID, username)
	if err != nil {
		return err
	}
	return nil
}
