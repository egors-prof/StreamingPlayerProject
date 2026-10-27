// Package streaming — выдача билетов на прослушивание и открытие аудиофайлов.
package streaming

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/authjwt"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/domain"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/port/driven"
)

type UseCase struct {
	tracks       driven.TrackStorage
	jwt          *authjwt.Manager
	musicDir     string
	ticketTTL    time.Duration
	publicPrefix string
}

func New(tracks driven.TrackStorage, jwt *authjwt.Manager, musicDir string, ticketTTL time.Duration, publicPrefix string) *UseCase {
	return &UseCase{tracks: tracks, jwt: jwt, musicDir: musicDir, ticketTTL: ticketTTL, publicPrefix: publicPrefix}
}

// IssueTicket выдаёт подписанную ссылку на трек. Билет действует только для
// одной песни и недолго: если ссылка утечёт, вред ограничен.
func (u *UseCase) IssueTicket(ctx context.Context, userID, songID int) (domain.Ticket, error) {
	if _, err := u.tracks.GetTrack(ctx, songID); err != nil {
		return domain.Ticket{}, err
	}
	token, claims, err := u.jwt.Issue(authjwt.Claims{
		UserID: userID,
		Type:   authjwt.TypeStream,
		SongID: songID,
	}, u.ticketTTL)
	if err != nil {
		return domain.Ticket{}, err
	}
	link := fmt.Sprintf("%s/stream/songs/%d?ticket=%s", u.publicPrefix, songID, url.QueryEscape(token))
	return domain.Ticket{URL: link, ExpiresAt: claims.ExpiresAt.Time}, nil
}

// OpenTrack проверяет билет и открывает файл. Путь к файлу берётся из БД по
// ID песни, а не из запроса пользователя, поэтому подставить "../" нельзя.
func (u *UseCase) OpenTrack(ctx context.Context, songID int, ticket string) (domain.TrackFile, error) {
	claims, err := u.jwt.Parse(ticket, authjwt.TypeStream)
	if err != nil || claims.SongID != songID {
		return domain.TrackFile{}, errs.ErrInvalidTicket
	}

	track, err := u.tracks.GetTrack(ctx, songID)
	if err != nil {
		return domain.TrackFile{}, err
	}

	// filepath.Base — вторая линия защиты: даже если в БД окажется путь
	// с каталогами, файл ищется строго внутри musicDir.
	path := filepath.Join(u.musicDir, filepath.Base(track.FileName))
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.TrackFile{}, fmt.Errorf("%w: file %s is missing", errs.ErrNotFound, track.FileName)
	}
	if err != nil {
		return domain.TrackFile{}, fmt.Errorf("open track: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return domain.TrackFile{}, fmt.Errorf("stat track: %w", err)
	}
	return domain.TrackFile{
		Content:     f,
		Name:        track.FileName,
		ContentType: track.ContentType,
		ModTime:     info.ModTime(),
	}, nil
}
