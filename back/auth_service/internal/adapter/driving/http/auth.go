package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/egors-prof/auth_service/internal/port/usecase"
	"github.com/egors-prof/streamingplayer/pkg/httpx"
	"github.com/gin-gonic/gin"
)

const refreshCookie = "refresh_token"

type SignUpRequest struct {
	FullName string `json:"full_name" example:"Egor Senatorov"`
	Username string `json:"username" example:"egor"`
	Password string `json:"password" example:"correct-horse-battery"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SessionResponse — ответ на вход и обновление сессии. Refresh-токена здесь
// нет: он приходит только в HttpOnly-cookie.
type SessionResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        UserResponse `json:"user"`
}

func toUserResponse(u domain.User) UserResponse {
	return UserResponse{ID: u.ID, Username: u.Username, FullName: u.FullName, Role: string(u.Role)}
}

// SignUp
// @Summary  Регистрация
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    request body SignUpRequest true "данные нового пользователя"
// @Success  201 {object} UserResponse
// @Failure  409 {object} httpx.ErrorResponse "логин занят"
// @Failure  422 {object} httpx.ErrorResponse "невалидные данные"
// @Failure  429 {object} httpx.ErrorResponse
// @Router   /auth/sign-up [post]
func (s *Server) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := s.registration.SignUp(c.Request.Context(), domain.SignUpInput{
		FullName: req.FullName, Username: req.Username, Password: req.Password,
	})
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

// SignIn
// @Summary  Вход
// @Description Возвращает access-токен в теле и ставит refresh-токен в HttpOnly-cookie.
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    request body SignInRequest true "логин и пароль"
// @Success  200 {object} SessionResponse
// @Failure  401 {object} httpx.ErrorResponse
// @Failure  429 {object} httpx.ErrorResponse
// @Router   /auth/sign-in [post]
func (s *Server) SignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	pair, err := s.session.SignIn(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		s.handleError(c, err)
		return
	}
	s.respondWithSession(c, pair)
}

// Refresh
// @Summary  Обновить сессию
// @Description Меняет refresh-токен из cookie на новую пару токенов (ротация).
// @Tags     Auth
// @Produce  json
// @Success  200 {object} SessionResponse
// @Failure  401 {object} httpx.ErrorResponse
// @Router   /auth/refresh [post]
func (s *Server) Refresh(c *gin.Context) {
	token, err := c.Cookie(refreshCookie)
	if err != nil || token == "" {
		s.handleError(c, errs.ErrInvalidToken)
		return
	}
	pair, err := s.session.Refresh(c.Request.Context(), token)
	if err != nil {
		// При гонке двух вкладок cookie не трогаем: в браузере уже лежит
		// новый токен, который только что получила другая вкладка.
		if !errors.Is(err, errs.ErrConcurrentRefresh) {
			s.clearRefreshCookie(c)
		}
		s.handleError(c, err)
		return
	}
	s.respondWithSession(c, pair)
}

// Logout
// @Summary  Выход
// @Tags     Auth
// @Success  204
// @Router   /auth/logout [post]
func (s *Server) Logout(c *gin.Context) {
	if token, err := c.Cookie(refreshCookie); err == nil && token != "" {
		if err := s.session.Logout(c.Request.Context(), token); err != nil {
			s.handleError(c, err)
			return
		}
	}
	s.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (s *Server) respondWithSession(c *gin.Context, pair usecase.TokenPair) {
	s.setRefreshCookie(c, pair.RefreshToken, int(time.Until(pair.RefreshExpiresAt).Seconds()))
	c.JSON(http.StatusOK, SessionResponse{
		AccessToken: pair.AccessToken,
		ExpiresAt:   pair.AccessExpiresAt,
		User:        toUserResponse(pair.User),
	})
}

// SameSite=Strict: браузер не отправит cookie с чужого сайта (защита от CSRF).
// Path ограничивает cookie эндпоинтами авторизации.
func (s *Server) setRefreshCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookie,
		Value:    value,
		Path:     s.cfg.CookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) clearRefreshCookie(c *gin.Context) {
	s.setRefreshCookie(c, "", -1)
}
