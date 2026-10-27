package dbstore

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type LikeStorage struct {
	db *sqlx.DB
}

func NewLikeStorage(db *sqlx.DB) *LikeStorage { return &LikeStorage{db: db} }

func (s *LikeStorage) AddLike(ctx context.Context, userID, songID int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO likes (user_id, song_id) VALUES ($1, $2)
		ON CONFLICT (user_id, song_id) DO NOTHING`, userID, songID)
	if err != nil {
		return fmt.Errorf("insert like: %w", err)
	}
	return nil
}

func (s *LikeStorage) RemoveLike(ctx context.Context, userID, songID int) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM likes WHERE user_id = $1 AND song_id = $2`, userID, songID)
	if err != nil {
		return fmt.Errorf("delete like: %w", err)
	}
	return nil
}

func (s *LikeStorage) LikedSongIDs(ctx context.Context, userID int) ([]int, error) {
	ids := make([]int, 0)
	err := s.db.SelectContext(ctx, &ids, `
		SELECT song_id FROM likes
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("select likes: %w", err)
	}
	return ids, nil
}
