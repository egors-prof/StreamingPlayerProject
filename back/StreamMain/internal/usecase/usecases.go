package usecase

import (
	"github.com/egors-prof/streaming/internal/adapter/driven/dbstore"
	"github.com/egors-prof/streaming/internal/config"

	"github.com/egors-prof/streaming/internal/port/usecase"
	"github.com/egors-prof/streaming/internal/usecase/get_path_by_title"
)

type UseCases struct {
	GetPathByTitle usecase.GetPathByTitleUseCase
}

func New(cfg *config.Config, store *dbstore.DBStore) *UseCases {
	return &UseCases{
		GetPathByTitle: getpathbytitle.New(cfg, store),
	}
}
