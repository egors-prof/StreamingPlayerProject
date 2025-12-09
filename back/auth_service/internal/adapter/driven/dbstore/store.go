package dbstore

import (
	"github.com/jmoiron/sqlx"
)

type DBStore struct {
	UserStorage *UserStorage
}

func New(db *sqlx.DB) *DBStore {
	return &DBStore{
		UserStorage: NewUserStorage(db),
	}
}
