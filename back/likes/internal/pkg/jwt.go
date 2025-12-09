package pkg

import (
	"fmt"

	"log"
	"os"

	"github.com/egors-prof/likes_service/internal/domain"
	"github.com/golang-jwt/jwt"
)

type CustomClaims struct {
	jwt.StandardClaims
	UserID    int         `json:"user_id"`
	Username  string      `json:"username"`
	Role      domain.Role `json:"role"`
	IsRefresh bool        `json:"isRefresh"`
}

func ParseToken(tokenString string) (int, string, bool, domain.Role, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {

			return nil, fmt.Errorf("unexpected signing method:%v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return 0, "", false, "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		log.Println("token String ,claims.UserID,claims.isRefresh", tokenString, claims.UserID, claims.IsRefresh)
		return claims.UserID, claims.Username, claims.IsRefresh, claims.Role, nil
	}
	return 0, "", false, "", fmt.Errorf("invalid token")
}
