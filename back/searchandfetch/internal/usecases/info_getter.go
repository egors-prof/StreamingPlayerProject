package usecases

import (
	"context"

	"github.com/egors-prof/searchService/internal/domain"
	"github.com/egors-prof/searchService/internal/repository"
)

type InfoGetter struct {
	repository *repository.Repository
}

func NewInfoGetter(repo *repository.Repository) *InfoGetter {
	return &InfoGetter{repository: repo}
}



func (i*InfoGetter) GetFirstSongs(ctx context.Context,quantity,offset int)([]domain.Song,error){
	dSongs,err:=i.repository.GetSongsInfo(ctx,quantity,offset)
	if err!=nil{
		return nil,err
	}
	return dSongs,nil
}
