package dbstore

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DBStore struct {
	LikeStorage *LikeStorage
}

func New(db *sqlx.DB) *DBStore {
	return &DBStore{
		LikeStorage: NewLikeStorage(db),
	}
}

func (ds *DBStore) SetLike(ctx context.Context, songID int, username string) error {
	err := ds.LikeStorage.SetLike(ctx, songID, username)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DBStore) RemoveLike(ctx context.Context, username string, songID int) error {
	err := ds.LikeStorage.RemoveLike(ctx, username, songID)
	if err != nil {
		return err
	}
	return nil
}
