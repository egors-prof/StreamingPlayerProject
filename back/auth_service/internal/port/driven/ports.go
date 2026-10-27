// Package driven — интерфейсы того, что нужно бизнес-логике от внешнего мира
// (БД, брокер сообщений). Реализации лежат в internal/adapter/driven.
package driven

import (
	"context"

	"github.com/egors-prof/auth_service/internal/domain"
)

type UserStorage interface {
	// CreateUser возвращает errs.ErrUsernameTaken, если логин занят.
	CreateUser(ctx context.Context, u domain.User) (int, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
}

type TokenStorage interface {
	SaveRefreshToken(ctx context.Context, t domain.RefreshToken) error
	// RotateRefreshToken в одной транзакции отзывает старый токен и сохраняет
	// новый. Если старый уже отозван — errs.ErrTokenReused, если его нет — errs.ErrNotFound.
	RotateRefreshToken(ctx context.Context, oldID string, next domain.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeAllUserTokens(ctx context.Context, userID int) error
}

type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, e domain.UserRegistered) error
}
