package pkg

import (
	"fmt"
	"github.com/egors-prof/streaming/internal/domain"
	"github.com/golang-jwt/jwt"
	"log"
	"os"
	"time"
)

type CustomClaims struct {
	jwt.StandardClaims
	UserID    int         `json:"user_id"`
	Username  string      `json:"username"`
	Role      domain.Role `json:"role"`
	IsRefresh bool        `json:"isRefresh"`
}

func GenerateToken(userId int, username string, ttl int, isRefresh bool, role domain.Role) (string, error) {

	if isRefresh {
		log.Println("generating refreshing token")

	} else {
		log.Println("generating access token")

	}
	claims := CustomClaims{StandardClaims: jwt.StandardClaims{},
		UserID:    userId,
		Username:  username,
		IsRefresh: isRefresh,
		Role:      role,
	}
	if isRefresh {
		claims.StandardClaims.ExpiresAt = int64(time.Duration(ttl) * 24 * time.Hour)
		log.Println("claims.StandardClaims.ExpiresAt (refresh token)", claims.StandardClaims.ExpiresAt)
	} else {
		claims.StandardClaims.ExpiresAt = int64(time.Duration(ttl) * 5 * time.Minute)
		log.Println("claims.StandardClaims.ExpiresAt (access token)", claims.StandardClaims.ExpiresAt)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	return tokenString, nil
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
