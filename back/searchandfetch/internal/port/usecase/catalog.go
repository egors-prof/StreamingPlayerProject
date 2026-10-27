package usecase

import (
	"context"

	"github.com/egors-prof/searchService/internal/domain"
)

type Catalog interface {
	List(ctx context.Context, afterID, limit int) (domain.Page, error)
	Search(ctx context.Context, query string, limit int) ([]domain.Song, error)
	GetByIDs(ctx context.Context, ids []int) ([]domain.Song, error)
	Get(ctx context.Context, id int) (domain.Song, error)
}
