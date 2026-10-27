// Package errs — доменные ошибки сервиса. HTTP-слой переводит их в коды ответа.
package errs

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
	// ErrTokenReused: refresh-токен, который уже обменяли, пришёл повторно.
	// Значит, его скопировали — отзываем все сессии пользователя.
	ErrTokenReused = errors.New("refresh token reuse detected")
	// ErrConcurrentRefresh: тот же токен только что обменяла другая вкладка.
	// Это не кража; клиенту достаточно повторить refresh с новым cookie.
	ErrConcurrentRefresh = errors.New("refresh token was just rotated, retry")
)
