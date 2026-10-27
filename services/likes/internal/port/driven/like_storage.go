package driven

import "context"

type LikeStorage interface {
	// AddLike идемпотентна: повторный лайк той же песни не создаёт дубль.
	AddLike(ctx context.Context, userID, songID int) error
	RemoveLike(ctx context.Context, userID, songID int) error
	// LikedSongIDs возвращает ID песен, новые лайки первыми.
	LikedSongIDs(ctx context.Context, userID int) ([]int, error)
}
