package errs

import "errors"

var (
	ErrNotfound                    = errors.New("not found")
	ErrBookNotFound                = errors.New("book not found")
	ErrProductNotfound             = errors.New("book not found")
	ErrInvalidBookID               = errors.New("invalid book id")
	ErrNoBookWithFollowingTitle    = errors.New("there is no book with following title")
	ErrInvalidRequestBody          = errors.New("invalid request body")
	ErrInvalidFieldValue           = errors.New("invalid field value")
	ErrInvalidProductName          = errors.New("invalid book name, min 4 symbols")
	ErrUsernameAlreadyExists       = errors.New("username already exists")
	ErrIncorrectUsernameOrPassword = errors.New("incorrect username or password")
	ErrInvalidToken                = errors.New("invalid token")
	ErrSomethingWentWrong          = errors.New("something went wrong")
	ErrNoFollowingAuthor           = errors.New("following author doesn't exists")
	ErrNoUsername                  = errors.New("there is no such username")
)
