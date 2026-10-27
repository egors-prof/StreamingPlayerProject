package errs

import "errors"

var (
	ErrNotFound     = errors.New("song not found")
	ErrInvalidInput = errors.New("invalid input")
)
