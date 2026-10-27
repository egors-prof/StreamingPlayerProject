// Package http — HTTP API лайков. Все ручки требуют access-токен.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/egors-prof/StreamingPlayerProject/pkg/authjwt"
	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/errs"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/port/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	likes usecase.Likes
	log   *slog.Logger
}

func New(trustedProxies []string, likes usecase.Likes, jwt *authjwt.Manager, log *slog.Logger) (http.Handler, error) {
	r, err := httpx.NewRouter(log, trustedProxies)
	if err != nil {
		return nil, err
	}
	s := &Server{likes: likes, log: log}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Middleware проверки токена подключён ко всей группе. В старой версии
	// он был написан, но не подключён ни к одному маршруту.
	g := r.Group("/likes", httpx.RequireAccessToken(jwt))
	g.GET("", s.List)
	// PUT и DELETE идемпотентны: повторный запрос даёт тот же результат.
	// Раньше лайк ставился через GET, а GET не должен менять данные:
	// браузеры и прокси повторяют и предзагружают такие запросы.
	g.PUT("/:songID", s.Like)
	g.DELETE("/:songID", s.Unlike)
	return r, nil
}

type LikesResponse struct {
	SongIDs []int `json:"song_ids"`
}

// List
// @Summary  Мои лайки
// @Tags     Likes
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} LikesResponse
// @Failure  401 {object} httpx.ErrorResponse
// @Router   /likes [get]
func (s *Server) List(c *gin.Context) {
	ids, err := s.likes.List(c.Request.Context(), httpx.ClaimsFrom(c).UserID)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, LikesResponse{SongIDs: ids})
}

// Like
// @Summary  Поставить лайк
// @Tags     Likes
// @Security BearerAuth
// @Param    songID path int true "ID песни"
// @Success  204
// @Failure  400 {object} httpx.ErrorResponse
// @Failure  401 {object} httpx.ErrorResponse
// @Router   /likes/{songID} [put]
func (s *Server) Like(c *gin.Context) {
	s.withSongID(c, s.likes.Like)
}

// Unlike
// @Summary  Убрать лайк
// @Tags     Likes
// @Security BearerAuth
// @Param    songID path int true "ID песни"
// @Success  204
// @Failure  400 {object} httpx.ErrorResponse
// @Failure  401 {object} httpx.ErrorResponse
// @Router   /likes/{songID} [delete]
func (s *Server) Unlike(c *gin.Context) {
	s.withSongID(c, s.likes.Unlike)
}

func (s *Server) withSongID(c *gin.Context, action func(ctx context.Context, userID, songID int) error) {
	songID, err := strconv.Atoi(c.Param("songID"))
	if err != nil {
		s.handleError(c, errs.ErrInvalidSongID)
		return
	}
	if err := action(c.Request.Context(), httpx.ClaimsFrom(c).UserID, songID); err != nil {
		s.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) handleError(c *gin.Context, err error) {
	if errors.Is(err, errs.ErrInvalidSongID) {
		httpx.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	s.log.Error("request failed", "error", err, "request_id", c.GetString("request_id"))
	httpx.Fail(c, http.StatusInternalServerError, "internal error")
}
