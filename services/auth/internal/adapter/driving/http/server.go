// Package http — входящий адаптер: HTTP API на Gin.
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/services/auth/internal/config"
	"github.com/egors-prof/StreamingPlayerProject/services/auth/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/auth/internal/port/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	registration usecase.Registration
	session      usecase.Session
	cfg          config.Auth
	log          *slog.Logger
}

func New(cfg config.Config, reg usecase.Registration, sess usecase.Session, log *slog.Logger) (http.Handler, error) {
	r, err := httpx.NewRouter(log, cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}
	s := &Server{registration: reg, session: sess, cfg: cfg.Auth, log: log}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := r.Group("/auth")
	// Защита от перебора паролей: не больше 10 попыток в минуту с одного IP.
	strict := httpx.RateLimit(10, time.Minute)
	auth.POST("/sign-up", strict, s.SignUp)
	auth.POST("/sign-in", strict, s.SignIn)
	auth.POST("/refresh", httpx.RateLimit(30, time.Minute), s.Refresh)
	auth.POST("/logout", s.Logout)
	return r, nil
}

// handleError переводит доменные ошибки в HTTP-коды. Неизвестные ошибки
// логируются, а клиенту уходит только "internal error": тексты ошибок БД
// не должны попадать наружу.
func (s *Server) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		httpx.Fail(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, errs.ErrUsernameTaken):
		httpx.Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, errs.ErrInvalidCredentials),
		errors.Is(err, errs.ErrInvalidToken),
		errors.Is(err, errs.ErrTokenReused),
		errors.Is(err, errs.ErrConcurrentRefresh):
		httpx.Fail(c, http.StatusUnauthorized, err.Error())
	default:
		s.log.Error("request failed", "error", err, "request_id", c.GetString("request_id"))
		httpx.Fail(c, http.StatusInternalServerError, "internal error")
	}
}
