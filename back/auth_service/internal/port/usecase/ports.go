// Package usecase — интерфейсы бизнес-сценариев, которые вызывает HTTP-слой.
package usecase

import (
	"context"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
)

type Registration interface {
	SignUp(ctx context.Context, in domain.SignUpInput) (domain.User, error)
}

// TokenPair — результат входа или обновления сессии.
type TokenPair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
	User             domain.User
}

type Session interface {
	SignIn(ctx context.Context, username, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}
