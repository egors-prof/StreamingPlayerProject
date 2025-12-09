package driven

import "context"

type LikeStorage interface {
	SetLike(ctx context.Context, songID int, username string) error
	RemoveLike(ctx context.Context, username string, songID int) error
}
