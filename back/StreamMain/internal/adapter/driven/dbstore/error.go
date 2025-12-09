package dbstore

import (
	"database/sql"
	"errors"

	"github.com/egors-prof/streaming/internal/errs"
)

func (s *SongStorage) translateError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errs.ErrNotfound
	default:
		return err
	}
}
