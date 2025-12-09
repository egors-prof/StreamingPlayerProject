package usecase

import "context"

type LikeRemover interface {
	RemoveLike(ctx context.Context, songID int, username string) error
}
