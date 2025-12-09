package usecase

import (
	"context"

	"github.com/egors-prof/streaming/internal/domain"
)

type GetPathByTitleUseCase interface {
	GetPathByTitle(ctx context.Context, title string) (*domain.Song, error)
}
