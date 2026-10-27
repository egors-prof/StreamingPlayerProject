package usecase

import "context"

type Likes interface {
	Like(ctx context.Context, userID, songID int) error
	Unlike(ctx context.Context, userID, songID int) error
	List(ctx context.Context, userID int) ([]int, error)
}
