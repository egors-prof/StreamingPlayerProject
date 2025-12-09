package dbstore

import (
	"database/sql"
	"errors"

	"github.com/egors-prof/auth_service/internal/errs"
)

func (u *UserStorage) translateError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errs.ErrNotfound
	default:
		return err
	}
}
