package usecase

import (
	"context"

	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/domain"
)

type Streaming interface {
	IssueTicket(ctx context.Context, userID, songID int) (domain.Ticket, error)
	OpenTrack(ctx context.Context, songID int, ticket string) (domain.TrackFile, error)
}
