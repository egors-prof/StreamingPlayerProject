package usecase

import "context"

type LikeSetter interface {
	SetLike(ctx context.Context, songID int, username string) error
}
