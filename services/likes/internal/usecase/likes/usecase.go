// Package likes — сценарии «лайкнуть», «убрать лайк», «мои лайки».
package likes

import (
	"context"

	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/port/driven"
)

type UseCase struct {
	storage driven.LikeStorage
}

func New(storage driven.LikeStorage) *UseCase { return &UseCase{storage: storage} }

// Существование песни здесь не проверяется: каталог — другой сервис с другой
// БД, внешнего ключа между ними нет. Варианты решения (синхронный запрос
// в каталог или события об удалении песен) описаны в README.
func (u *UseCase) Like(ctx context.Context, userID, songID int) error {
	if songID < 1 {
		return errs.ErrInvalidSongID
	}
	return u.storage.AddLike(ctx, userID, songID)
}

func (u *UseCase) Unlike(ctx context.Context, userID, songID int) error {
	if songID < 1 {
		return errs.ErrInvalidSongID
	}
	return u.storage.RemoveLike(ctx, userID, songID)
}

func (u *UseCase) List(ctx context.Context, userID int) ([]int, error) {
	return u.storage.LikedSongIDs(ctx, userID)
}
