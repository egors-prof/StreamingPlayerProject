// Package session — вход, обновление пары токенов и выход.
package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/egors-prof/auth_service/internal/port/driven"
	"github.com/egors-prof/auth_service/internal/port/usecase"
	"github.com/egors-prof/streamingplayer/pkg/authjwt"
	"golang.org/x/crypto/bcrypt"
)

// reuseGrace — сколько секунд после ротации повтор старого refresh-токена
// считается не кражей, а гонкой: две вкладки одновременно обновили сессию
// одним и тем же cookie. Такой запрос получает ErrConcurrentRefresh без отзыва
// всех сессий, а фронт повторяет refresh уже с новым cookie.
const reuseGrace = 10 * time.Second

type UseCase struct {
	now        func() time.Time
	users      driven.UserStorage
	tokens     driven.TokenStorage
	jwt        *authjwt.Manager
	accessTTL  time.Duration
	refreshTTL time.Duration
	dummyHash  []byte
}

func New(users driven.UserStorage, tokens driven.TokenStorage, jwt *authjwt.Manager, accessTTL, refreshTTL time.Duration) *UseCase {
	// Хеш-пустышка нужен, чтобы вход с несуществующим логином занимал
	// столько же времени, сколько с неверным паролем. Иначе по времени
	// ответа можно узнать, какие логины зарегистрированы.
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &UseCase{
		now:   time.Now,
		users: users, tokens: tokens, jwt: jwt,
		accessTTL: accessTTL, refreshTTL: refreshTTL, dummyHash: dummy,
	}
}

// SignIn проверяет логин и пароль. На «нет такого пользователя» и «неверный
// пароль» отвечает одной и той же ошибкой.
func (u *UseCase) SignIn(ctx context.Context, username, password string) (usecase.TokenPair, error) {
	user, err := u.users.GetUserByUsername(ctx, username)
	switch {
	case errors.Is(err, errs.ErrNotFound):
		_ = bcrypt.CompareHashAndPassword(u.dummyHash, []byte(password))
		return usecase.TokenPair{}, errs.ErrInvalidCredentials
	case err != nil:
		return usecase.TokenPair{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return usecase.TokenPair{}, errs.ErrInvalidCredentials
	}

	pair, refresh, err := u.issuePair(user)
	if err != nil {
		return usecase.TokenPair{}, err
	}
	if err := u.tokens.SaveRefreshToken(ctx, refresh); err != nil {
		return usecase.TokenPair{}, fmt.Errorf("save refresh token: %w", err)
	}
	return pair, nil
}

// Refresh меняет refresh-токен на новую пару (ротация). Каждый refresh-токен
// одноразовый: повторное использование означает кражу, и тогда отзываются
// все сессии пользователя.
func (u *UseCase) Refresh(ctx context.Context, refreshToken string) (usecase.TokenPair, error) {
	claims, err := u.jwt.Parse(refreshToken, authjwt.TypeRefresh)
	if err != nil {
		return usecase.TokenPair{}, errs.ErrInvalidToken
	}

	// Данные пользователя берём из БД, а не из старого токена: роль могла измениться.
	user, err := u.users.GetUserByID(ctx, claims.UserID)
	if errors.Is(err, errs.ErrNotFound) {
		return usecase.TokenPair{}, errs.ErrInvalidToken
	}
	if err != nil {
		return usecase.TokenPair{}, err
	}

	pair, next, err := u.issuePair(user)
	if err != nil {
		return usecase.TokenPair{}, err
	}
	err = u.tokens.RotateRefreshToken(ctx, claims.ID, next)
	switch {
	case errors.Is(err, errs.ErrTokenReused):
		rotatedAt, rErr := u.tokens.RefreshTokenRotatedAt(ctx, claims.ID)
		if rErr == nil && !rotatedAt.IsZero() && u.now().Sub(rotatedAt) < reuseGrace {
			return usecase.TokenPair{}, errs.ErrConcurrentRefresh
		}
		if revokeErr := u.tokens.RevokeAllUserTokens(ctx, user.ID); revokeErr != nil {
			return usecase.TokenPair{}, revokeErr
		}
		return usecase.TokenPair{}, errs.ErrTokenReused
	case errors.Is(err, errs.ErrNotFound):
		return usecase.TokenPair{}, errs.ErrInvalidToken
	case err != nil:
		return usecase.TokenPair{}, err
	}
	return pair, nil
}

// Logout отзывает refresh-токен. Access-токен продолжит работать до истечения
// (до 15 минут) — это плата за проверку access-токена без похода в БД.
func (u *UseCase) Logout(ctx context.Context, refreshToken string) error {
	claims, err := u.jwt.Parse(refreshToken, authjwt.TypeRefresh)
	if err != nil {
		return nil // токен и так недействителен — выходить не из чего
	}
	return u.tokens.RevokeRefreshToken(ctx, claims.ID)
}

func (u *UseCase) issuePair(user domain.User) (usecase.TokenPair, domain.RefreshToken, error) {
	base := authjwt.Claims{UserID: user.ID, Username: user.Username, Role: authjwt.Role(user.Role)}

	access := base
	access.Type = authjwt.TypeAccess
	accessToken, accessClaims, err := u.jwt.Issue(access, u.accessTTL)
	if err != nil {
		return usecase.TokenPair{}, domain.RefreshToken{}, err
	}

	refresh := authjwt.Claims{UserID: user.ID, Type: authjwt.TypeRefresh}
	refreshToken, refreshClaims, err := u.jwt.Issue(refresh, u.refreshTTL)
	if err != nil {
		return usecase.TokenPair{}, domain.RefreshToken{}, err
	}

	pair := usecase.TokenPair{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: refreshClaims.ExpiresAt.Time,
		User:             user,
	}
	record := domain.RefreshToken{ID: refreshClaims.ID, UserID: user.ID, ExpiresAt: refreshClaims.ExpiresAt.Time}
	return pair, record, nil
}
