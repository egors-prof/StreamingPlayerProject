package config

import (
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/pgdb"
)

type Config struct {
	HTTPPort       string   `env:"HTTP_PORT, default=9999"`
	LogLevel       string   `env:"LOG_LEVEL, default=info"`
	TrustedProxies []string `env:"TRUSTED_PROXIES"`
	JWTSecret      string   `env:"AUTH_JWT_SECRET, required"`

	Postgres pgdb.Config `env:", prefix=POSTGRES_"`
	Stream   Stream      `env:", prefix=STREAM_"`
}

type Stream struct {
	// Каталог с аудиофайлами.
	MusicDir string `env:"MUSIC_DIR, default=./music"`
	// Сколько стримов сервер отдаёт одновременно.
	MaxConcurrent int `env:"MAX_CONCURRENT, default=50"`
	// Сколько запрос ждёт свободного слота, прежде чем получить 503.
	QueueTimeout time.Duration `env:"QUEUE_TIMEOUT, default=3s"`
	// Срок жизни билета на прослушивание (подписанной ссылки на трек).
	TicketTTL time.Duration `env:"TICKET_TTL, default=1h"`
	// Публичный префикс пути, по которому фронт обращается к сервису через шлюз.
	PublicPrefix string `env:"PUBLIC_PREFIX, default=/api"`
}
