package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/egors-prof/auth_service/internal/errs"
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID           int
	FullName     string
	Username     string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SignUpInput — данные формы регистрации.
type SignUpInput struct {
	FullName string
	Username string
	Password string
}

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)

// Validate проверяет данные регистрации. bcrypt учитывает только первые
// 72 байта пароля, поэтому более длинные пароли отклоняем явно.
func (in *SignUpInput) Validate() error {
	in.FullName = strings.TrimSpace(in.FullName)
	in.Username = strings.TrimSpace(in.Username)

	switch n := utf8.RuneCountInString(in.FullName); {
	case n == 0:
		return fmt.Errorf("%w: full name is required", errs.ErrInvalidInput)
	case n > 100:
		return fmt.Errorf("%w: full name must be at most 100 characters", errs.ErrInvalidInput)
	}
	if !usernameRe.MatchString(in.Username) {
		return fmt.Errorf("%w: username must be 3-32 characters: letters, digits, '_', '.', '-'", errs.ErrInvalidInput)
	}
	if len(in.Password) < 8 || len(in.Password) > 72 {
		return fmt.Errorf("%w: password must be 8-72 bytes long", errs.ErrInvalidInput)
	}
	return nil
}

// RefreshToken — запись о выданном refresh-токене. Сам токен не храним,
// только его идентификатор (jti): этого достаточно для отзыва.
type RefreshToken struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}

// UserRegistered — событие, которое публикуется в RabbitMQ после регистрации.
type UserRegistered struct {
	UserID     int       `json:"user_id"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	OccurredAt time.Time `json:"occurred_at"`
}
