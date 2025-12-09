package likeRemover

import (
	"context"

	"github.com/egors-prof/likes_service/internal/config"
	"github.com/egors-prof/likes_service/internal/port/driven"
)

type UseCase struct {
	cfg         *config.Config
	LikeStorage driven.LikeStorage
}

func New(cfg *config.Config, LikeStorage driven.LikeStorage) *UseCase {
	return &UseCase{
		cfg:         cfg,
		LikeStorage: LikeStorage,
	}
}

func (u *UseCase) RemoveLike(ctx context.Context, songID int, username string) error {
	err := u.LikeStorage.RemoveLike(ctx, username, songID)
	if err != nil {
		return err
	}
	return nil
}
