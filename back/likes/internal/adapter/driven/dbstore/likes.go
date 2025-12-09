package dbstore

import (
	"context"
	"time"

	"github.com/egors-prof/likes_service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type LikeStorage struct {
	db *sqlx.DB
}

func NewLikeStorage(db *sqlx.DB) *LikeStorage {
	return &LikeStorage{
		db: db,
	}
}

type Like struct {
	ID        int       `db:"id"`
	Username  int       `db:"username"`
	SongID    int       `db:"song_id"`
	CreatedAt time.Time `db:"created_at"`
}

func (rl *Like) FromRepositoryToDomain() *domain.Like {
	return &domain.Like{
		ID:        rl.ID,
		UserID:    rl.Username,
		SongID:    rl.SongID,
		CreatedAt: rl.CreatedAt,
	}
}

func (ls *LikeStorage) SetLike(ctx context.Context, songID int, username string) error {
	_, err := ls.db.ExecContext(ctx, `
		INSERT INTO likes (username, song_id) VALUES ($1, $2)`, username, songID)
	if err != nil {
		return ls.translateError(err)
	}
	return nil

}

func (ls *LikeStorage) RemoveLike(ctx context.Context, username string, songID int) error {
	_, err := ls.db.ExecContext(ctx, `
	DELETE FROM likes WHERE username = $1 AND song_id = $2
`, username, songID)
	if err != nil {
		return ls.translateError(err)
	}
	return nil
}
