package usecases

import (
	"context"

	"github.com/egors-prof/searchService/internal/domain"
	"github.com/egors-prof/searchService/internal/repository"
)

type Searcher struct {
	repository *repository.Repository
}

func NewSearcher(repository *repository.Repository) *Searcher {
	return &Searcher{
		repository: repository,
	}
}

func (s *Searcher) GetSearch(ctx context.Context, search string) ([]domain.Song, error) {
	dSongs, err := s.repository.GetSearch(ctx, search)
	if err != nil {
		return nil, err
	}
	return dSongs, nil
}
