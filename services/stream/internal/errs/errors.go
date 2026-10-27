package errs

import "errors"

var (
	ErrNotFound      = errors.New("track not found")
	ErrInvalidTicket = errors.New("invalid or expired stream ticket")
)
