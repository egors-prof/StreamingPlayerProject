package dbstore

import (
	"database/sql"
	"errors"

	"github.com/egors-prof/likes_service/internal/errs"
)

func (ls *LikeStorage) translateError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errs.ErrNotfound
	default:
		return err
	}
}
