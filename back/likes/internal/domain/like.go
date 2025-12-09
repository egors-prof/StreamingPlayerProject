package domain

import "time"

type Like struct {
	ID        int
	UserID    int
	SongID    int
	CreatedAt time.Time
}
