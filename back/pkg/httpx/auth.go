package httpx

import (
	"errors"
	"net/http"

	"github.com/egors-prof/streamingplayer/pkg/authjwt"
	"github.com/gin-gonic/gin"
)

const claimsKey = "auth_claims"

// RequireAccessToken пропускает запрос дальше только с валидным access-токеном.
// Refresh- и stream-токены здесь отклоняются по полю typ.
func RequireAccessToken(m *authjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := authjwt.BearerToken(c.GetHeader("Authorization"))
		if err != nil {
			Fail(c, http.StatusUnauthorized, err.Error())
			return
		}
		claims, err := m.Parse(token, authjwt.TypeAccess)
		if err != nil {
			msg := "invalid token"
			if errors.Is(err, authjwt.ErrExpiredToken) {
				msg = "token expired" // фронт по этому ответу понимает, что нужен refresh
			}
			Fail(c, http.StatusUnauthorized, msg)
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// ClaimsFrom возвращает claims, сохранённые RequireAccessToken.
func ClaimsFrom(c *gin.Context) *authjwt.Claims {
	v, ok := c.Get(claimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*authjwt.Claims)
	return claims
}
