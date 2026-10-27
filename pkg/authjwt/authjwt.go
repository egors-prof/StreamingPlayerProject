// Package authjwt выпускает и проверяет JWT для всех сервисов проекта.
//
// Раньше каждый сервис держал свою копию этого кода, и копии разошлись:
// сервис авторизации писал флаг refresh в поле "is_refresh", а остальные сервисы
// читали "isRefresh". Из-за этого refresh-токен принимался вместо access.
// Теперь структура claims одна на всех, а тип токена проверяется явно.
package authjwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType различает назначение токена. Токен одного типа нельзя
// использовать вместо другого.
type TokenType string

const (
	TypeAccess  TokenType = "access"  // короткий, отправляется с каждым запросом
	TypeRefresh TokenType = "refresh" // длинный, только для /auth/refresh
	TypeStream  TokenType = "stream"  // билет на прослушивание одного трека
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

const issuer = "streamingplayer-auth"

// MinSecretLength — минимальная длина секрета для HS256.
// Короткий секрет можно подобрать перебором по любому перехваченному токену.
const MinSecretLength = 32

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token expired")
	ErrWrongTokenType = errors.New("wrong token type")
	ErrNoBearerToken  = errors.New("missing bearer token")
)

// Claims — полезная нагрузка всех токенов проекта.
type Claims struct {
	jwt.RegisteredClaims
	UserID   int       `json:"uid"`
	Username string    `json:"username,omitempty"`
	Role     Role      `json:"role,omitempty"`
	Type     TokenType `json:"typ"`
	SongID   int       `json:"song_id,omitempty"` // только для TypeStream
}

// Manager подписывает и проверяет токены общим секретом (HS256).
type Manager struct {
	secret []byte
	now    func() time.Time
}

func NewManager(secret string) (*Manager, error) {
	if len(secret) < MinSecretLength {
		return nil, fmt.Errorf("JWT secret must be at least %d characters", MinSecretLength)
	}
	return &Manager{secret: []byte(secret), now: time.Now}, nil
}

// Issue подписывает токен со сроком жизни ttl. Возвращает строку токена и
// итоговые claims (в них заполнены ID, IssuedAt и ExpiresAt).
func (m *Manager) Issue(c Claims, ttl time.Duration) (string, Claims, error) {
	if c.Type == "" {
		return "", Claims{}, errors.New("token type is required")
	}
	now := m.now()
	id, err := newTokenID()
	if err != nil {
		return "", Claims{}, err
	}
	c.RegisteredClaims = jwt.RegisteredClaims{
		ID:       id,
		Issuer:   issuer,
		Subject:  fmt.Sprint(c.UserID),
		IssuedAt: jwt.NewNumericDate(now),
		// Главная ошибка старой версии: в ExpiresAt записывалась длительность
		// в наносекундах вместо момента времени, и токены жили тысячи лет.
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(m.secret)
	if err != nil {
		return "", Claims{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, c, nil
}

// Parse проверяет подпись, срок жизни, издателя и тип токена.
func (m *Manager) Parse(token string, want TokenType) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), // защита от alg=none и подмены алгоритма
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(issuer),
		jwt.WithTimeFunc(m.now),
	)
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return nil, ErrExpiredToken
	case err != nil:
		return nil, ErrInvalidToken
	case c.Type != want:
		return nil, ErrWrongTokenType
	case c.UserID <= 0:
		return nil, ErrInvalidToken
	}
	return &c, nil
}

// BearerToken достаёт токен из заголовка "Authorization: Bearer <token>".
func BearerToken(header string) (string, error) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", ErrNoBearerToken
	}
	return strings.TrimSpace(token), nil
}

func newTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
