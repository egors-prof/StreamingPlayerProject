// Package http — HTTP API стриминга.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/egors-prof/StreamingPlayerProject/pkg/authjwt"
	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/limiter"
	"github.com/egors-prof/StreamingPlayerProject/services/stream/internal/port/usecase"
	"github.com/gin-gonic/gin"
)

type Server struct {
	streaming usecase.Streaming
	limiter   *limiter.Limiter
	log       *slog.Logger
}

func New(trustedProxies []string, streaming usecase.Streaming, lim *limiter.Limiter, jwt *authjwt.Manager, log *slog.Logger) (http.Handler, error) {
	r, err := httpx.NewRouter(log, trustedProxies)
	if err != nil {
		return nil, err
	}
	s := &Server{streaming: streaming, limiter: lim, log: log}

	g := r.Group("/stream")
	g.POST("/tickets", httpx.RequireAccessToken(jwt), s.IssueTicket)
	// Сам файл защищён билетом в query, а не заголовком: <audio> не умеет
	// слать Authorization. Лимитер стоит только на отдаче аудио.
	g.GET("/songs/:id", s.limit, s.StreamSong)
	g.HEAD("/songs/:id", s.limit, s.StreamSong)
	g.GET("/stats", s.Stats)
	return r, nil
}

type TicketRequest struct {
	SongID int `json:"song_id" binding:"required,min=1"`
}

type TicketResponse struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

// IssueTicket: POST /stream/tickets {"song_id": 3} → ссылка для <audio src>.
func (s *Server) IssueTicket(c *gin.Context) {
	var req TicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, http.StatusBadRequest, "song_id is required")
		return
	}
	claims := httpx.ClaimsFrom(c)
	ticket, err := s.streaming.IssueTicket(c.Request.Context(), claims.UserID, req.SongID)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, TicketResponse{URL: ticket.URL, ExpiresAt: ticket.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z")})
}

// StreamSong отдаёт аудиофайл через http.ServeContent. Он сам обрабатывает
// заголовок Range и отвечает 206 Partial Content с нужным куском файла.
// Поэтому браузер начинает играть сразу, не скачивая файл целиком, а при
// перемотке запрашивает нужный байтовый диапазон.
func (s *Server) StreamSong(c *gin.Context) {
	songID, err := strconv.Atoi(c.Param("id"))
	if err != nil || songID < 1 {
		httpx.Fail(c, http.StatusBadRequest, "invalid song id")
		return
	}
	track, err := s.streaming.OpenTrack(c.Request.Context(), songID, c.Query("ticket"))
	if err != nil {
		s.handleError(c, err)
		return
	}
	defer track.Content.Close()

	c.Header("Content-Type", track.ContentType)
	c.Header("Cache-Control", "private, max-age=3600")
	http.ServeContent(c.Writer, c.Request, track.Name, track.ModTime, track.Content)
}

// limit занимает слот лимитера на всё время отдачи файла.
func (s *Server) limit(c *gin.Context) {
	release, err := s.limiter.Acquire(c.Request.Context())
	switch {
	case errors.Is(err, limiter.ErrBusy):
		c.Header("Retry-After", "5")
		httpx.Fail(c, http.StatusServiceUnavailable, "too many active streams, try again later")
		return
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		c.Abort() // клиент ушёл сам — отвечать некому
		return
	case err != nil:
		s.handleError(c, err)
		return
	}
	defer release()
	c.Next()
}

// Stats: GET /stream/stats — сколько стримов идёт сейчас и сколько отклонено.
func (s *Server) Stats(c *gin.Context) {
	c.JSON(http.StatusOK, s.limiter.Stats())
}

func (s *Server) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidTicket):
		httpx.Fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errs.ErrNotFound):
		httpx.Fail(c, http.StatusNotFound, "track not found")
	default:
		s.log.Error("request failed", "error", err, "request_id", c.GetString("request_id"))
		httpx.Fail(c, http.StatusInternalServerError, "internal error")
	}
}
