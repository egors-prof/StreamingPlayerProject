package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/egors-prof/likes_service/internal/pkg"
	"github.com/gin-gonic/gin"
)

type Like struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	SongID    int       `json:"song_id" db:"song_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LikeRequest struct {
	SongID int `json:"song_id" validate:"required,min=1"`
}

type LikeResponse struct {
	Success bool `json:"success"`
	Liked   bool `json:"liked"` // Текущее состояние

}

type UserLikesResponse struct {
	SongsID []int `json:"songs_id"`
	Total   int   `json:"total"`
}

// SetLike
// @Summary Поставить лайк песне
// @Description Добавляет лайк текущего пользователя к указанной песне.
// @Description Если лайк уже существует, операция игнорируется.
// @Tags likes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer токен авторизации в формате 'Bearer {token}'"
// @Param s query int true "ID песни" minimum(1)
// @Success 200 {object} LikeResponse
// @Failure 400 {object} CommonError
// @Failure 401 {object} CommonError
// @Failure 422 {object} CommonError
// @Failure 404 {object} CommonError
// @Failure 500 {object} CommonError
// @Router /like [get]
func (s *Server) SetLike(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	songIDStr := c.Query("s")
	songID, err := strconv.Atoi(songIDStr)
	if err != nil {
		s.handleError(c, err)
		return
	}
	userId, username, isRefresh, role, err := pkg.ParseToken(accessToken)
	if err != nil {
		s.handleError(c, err)
		return
	}
	fmt.Println(userId, username, isRefresh, role)
	err = s.uc.LikeSetter.SetLike(c, songID, username)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, LikeResponse{
		Success: true,
		Liked:   true,
	})
}

func (s *Server) RemoveLike(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	songIDStr := c.Query("s")
	songID, err := strconv.Atoi(songIDStr)
	if err != nil {
		s.handleError(c, err)
		return
	}
	userId, username, isRefresh, role, err := pkg.ParseToken(accessToken)
	fmt.Println(userId, username, isRefresh, role)
	if err != nil {
		s.handleError(c, err)
	}
	err = s.uc.LikeRemover.RemoveLike(c, songID, username)
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, LikeResponse{
		Success: true,
		Liked:   false,
	})
}
