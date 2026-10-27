// Package httpx содержит общий HTTP-код сервисов: запуск сервера с graceful
// shutdown, middleware и единый формат ошибок.
package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NewRouter создаёт gin.Engine с базовыми middleware. trustedProxies — сети,
// чьему заголовку X-Forwarded-For можно верить (например, nginx-шлюз).
func NewRouter(log *slog.Logger, trustedProxies []string) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("trusted proxies: %w", err)
	}
	r.Use(RequestID(), AccessLog(log), Recovery(log))
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ping": "pong"}) })
	return r, nil
}

// NewServer задаёт таймауты. WriteTimeout не ставим: стриминг аудио отдаёт
// ответ дольше любого разумного лимита, длительность ограничивают иначе.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Run запускает сервер и останавливает его, когда отменяется ctx.
// Shutdown перестаёт принимать новые соединения и ждёт текущие запросы.
func Run(ctx context.Context, srv *http.Server, log *slog.Logger, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("http server started", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	return nil
}
