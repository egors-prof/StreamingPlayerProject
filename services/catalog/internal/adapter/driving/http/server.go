// Package http — публичное API каталога (без авторизации).
package http

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/domain"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/port/usecase"
	"github.com/gin-gonic/gin"
)

type Server struct {
	catalog usecase.Catalog
	log     *slog.Logger
}

// CORS здесь больше не нужен: фронт и API отдаются с одного адреса через
// nginx-шлюз. Раньше было три CORS-middleware подряд, а сочетание
// AllowOrigins "*" и AllowCredentials true браузеры запрещают.
func New(trustedProxies []string, catalog usecase.Catalog, log *slog.Logger) (http.Handler, error) {
	r, err := httpx.NewRouter(log, trustedProxies)
	if err != nil {
		return nil, err
	}
	s := &Server{catalog: catalog, log: log}

	g := r.Group("/songs")
	g.GET("", s.List)
	g.GET("/search", s.Search)
	g.GET("/batch", s.Batch)
	g.GET("/:id", s.Get)
	return r, nil
}

type SongResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Album       string  `json:"album"`
	CoverURL    *string `json:"cover_url"`
	DurationSec int     `json:"duration_sec"`
}

type ListResponse struct {
	Items       []SongResponse `json:"items"`
	NextAfterID *int           `json:"next_after_id"`
}

func toResponse(songs []domain.Song) []SongResponse {
	out := make([]SongResponse, 0, len(songs))
	for _, s := range songs {
		resp := SongResponse{ID: s.ID, Title: s.Title, Artist: s.Artist, Album: s.Album, DurationSec: s.DurationSec}
		if s.CoverPath != "" {
			cover := s.CoverPath
			resp.CoverURL = &cover
		}
		out = append(out, resp)
	}
	return out
}

// List: GET /songs?limit=20&after_id=0
func (s *Server) List(c *gin.Context) {
	afterID, err1 := queryInt(c, "after_id")
	limit, err2 := queryInt(c, "limit")
	if err := errors.Join(err1, err2); err != nil {
		s.handleError(c, err)
		return
	}
	page, err := s.catalog.List(c.Request.Context(), afterID, limit)
	if err != nil {
		s.handleError(c, err)
		return
	}
	resp := ListResponse{Items: toResponse(page.Songs)}
	if page.NextAfterID > 0 {
		resp.NextAfterID = &page.NextAfterID
	}
	c.JSON(http.StatusOK, resp)
}

// Search: GET /songs/search?q=neon&limit=20
func (s *Server) Search(c *gin.Context) {
	limit, err := queryInt(c, "limit")
	if err != nil {
		s.handleError(c, err)
		return
	}
	songs, err := s.catalog.Search(c.Request.Context(), c.Query("q"), limit)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: toResponse(songs)})
}

// Batch: GET /songs/batch?ids=3,1,2 — песни в порядке переданных ID.
func (s *Server) Batch(c *gin.Context) {
	var ids []int
	if raw := strings.TrimSpace(c.Query("ids")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			id, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || id < 1 {
				s.handleError(c, errs.ErrInvalidInput)
				return
			}
			ids = append(ids, id)
		}
	}
	songs, err := s.catalog.GetByIDs(c.Request.Context(), ids)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: toResponse(songs)})
}

// Get: GET /songs/:id
func (s *Server) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		s.handleError(c, errs.ErrInvalidInput)
		return
	}
	song, err := s.catalog.Get(c.Request.Context(), id)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse([]domain.Song{song})[0])
}

func queryInt(c *gin.Context, key string) (int, error) {
	raw := c.Query(key)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errs.ErrInvalidInput
	}
	return n, nil
}

// handleError: раньше при ошибке обработчик просто выходил, и клиент
// получал пустой ответ 200. Теперь у каждой ошибки есть код и тело.
func (s *Server) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidInput):
		httpx.Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, errs.ErrNotFound):
		httpx.Fail(c, http.StatusNotFound, err.Error())
	default:
		s.log.Error("request failed", "error", err, "request_id", c.GetString("request_id"))
		httpx.Fail(c, http.StatusInternalServerError, "internal error")
	}
}
