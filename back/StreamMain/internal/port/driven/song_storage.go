package driven

import (
	"context"

	"github.com/egors-prof/streaming/internal/domain"
)

type SongStorage interface {
	GetPathByTitle(ctx context.Context, title string) (*domain.Song, error)
}
