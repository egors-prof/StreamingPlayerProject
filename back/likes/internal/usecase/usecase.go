package usecase

import (
	"github.com/egors-prof/likes_service/internal/config"
	"github.com/egors-prof/likes_service/internal/port/driven"
	"github.com/egors-prof/likes_service/internal/usecase/likeRemover"
	"github.com/egors-prof/likes_service/internal/usecase/likeSetter"
)

type UseCase struct {
	LikeSetter  likeSetter.UseCase
	LikeRemover likeRemover.UseCase
}

func New(cfg *config.Config, ls driven.LikeStorage) *UseCase {
	return &UseCase{
		LikeSetter:  *likeSetter.New(cfg, ls),
		LikeRemover: *likeRemover.New(cfg, ls),
	}
}
