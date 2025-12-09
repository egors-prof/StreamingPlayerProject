package usecases

import "github.com/egors-prof/searchService/internal/repository"

type UseCases struct {
	InfoGetter *InfoGetter
	Searcher   *Searcher
}

func NewUseCases(repo *repository.Repository) *UseCases {
	return &UseCases{
		InfoGetter: NewInfoGetter(repo),
		Searcher:   NewSearcher(repo),
	}

}
