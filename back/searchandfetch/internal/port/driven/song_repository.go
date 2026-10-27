package driven

import (
	"context"

	"github.com/egors-prof/searchService/internal/domain"
)

type SongRepository interface {
	// List возвращает до limit песен с id > afterID по возрастанию id.
	List(ctx context.Context, afterID, limit int) ([]domain.Song, error)
	Search(ctx context.Context, query string, limit int) ([]domain.Song, error)
	GetByIDs(ctx context.Context, ids []int) ([]domain.Song, error)
}
