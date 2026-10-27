package domain

import (
	"io"
	"time"
)

// Track — аудиофайл, привязанный к песне из каталога (song_id).
type Track struct {
	SongID      int
	FileName    string
	ContentType string
}

// Ticket — подписанная ссылка на прослушивание одного трека одним
// пользователем. Нужна потому, что тег <audio> не умеет отправлять заголовок
// Authorization: токен приходится класть в URL, и это должен быть
// короткоживущий токен с узкими правами, а не access-токен.
type Ticket struct {
	URL       string
	ExpiresAt time.Time
}

// TrackFile — открытый файл, готовый к отдаче с поддержкой Range.
type TrackFile struct {
	Content     io.ReadSeekCloser
	Name        string
	ContentType string
	ModTime     time.Time
}
