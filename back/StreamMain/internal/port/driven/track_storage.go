package driven

import (
	"context"

	"github.com/egors-prof/streaming/internal/domain"
)

type TrackStorage interface {
	// GetTrack возвращает errs.ErrNotFound, если для песни нет аудиофайла.
	GetTrack(ctx context.Context, songID int) (domain.Track, error)
}
