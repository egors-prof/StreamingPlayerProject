// Package registration — сценарий регистрации пользователя.
package registration

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/port/driven"
	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	users     driven.UserStorage
	publisher driven.EventPublisher
	log       *slog.Logger
	hashCost  int
}

func New(users driven.UserStorage, publisher driven.EventPublisher, log *slog.Logger) *UseCase {
	return &UseCase{users: users, publisher: publisher, log: log, hashCost: bcrypt.DefaultCost}
}

// SignUp создаёт пользователя. Проверку «логин свободен?» не делаем заранее:
// между проверкой и INSERT другой запрос мог занять логин (гонка
// check-then-act). Вместо этого полагаемся на UNIQUE в БД.
func (u *UseCase) SignUp(ctx context.Context, in domain.SignUpInput) (domain.User, error) {
	if err := in.Validate(); err != nil {
		return domain.User{}, err
	}

	// bcrypt: медленный хеш со встроенной солью. SHA-256 из прошлой версии
	// слишком быстрый — видеокарта перебирает миллиарды паролей в секунду.
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), u.hashCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		FullName:     in.FullName,
		Username:     in.Username,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}
	user.ID, err = u.users.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	// Событие публикуется после коммита в БД. Если брокер недоступен,
	// регистрация не откатывается, событие теряется. Надёжное решение —
	// transactional outbox (см. README).
	event := domain.UserRegistered{
		UserID:     user.ID,
		Username:   user.Username,
		FullName:   user.FullName,
		OccurredAt: time.Now().UTC(),
	}
	if err := u.publisher.PublishUserRegistered(ctx, event); err != nil {
		u.log.Error("publish user.registered failed", "user_id", user.ID, "error", err)
	}
	return user, nil
}
