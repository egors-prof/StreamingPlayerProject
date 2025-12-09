package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/egors-prof/likes_service/internal/pkg"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	userIDCtx           = "userID"
	userRoleCtx         = "userRole"
	UsernameCtx         = "Username"
)

func (s *Server) checkUserAuthentication(c *gin.Context) {
	token := c.Query("token")
	fmt.Println("middleware")
	fmt.Println(token)
	fmt.Println(c.GetHeader(authorizationHeader))
	token, err := s.extractTokenFromHeader(c, authorizationHeader)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, CommonError{Error: err.Error()})
		return
	}

	userID, username, isRefresh, userRole, err := pkg.ParseToken(token)
	if err != nil {
		log.Println("error!!!!!!!!!!")
		c.AbortWithStatusJSON(http.StatusUnauthorized, CommonError{Error: err.Error()})
		return
	}

	if isRefresh {
		c.AbortWithStatusJSON(http.StatusUnauthorized, CommonError{Error: "inappropriate token"})
		return
	}

	c.Set(userIDCtx, userID)
	c.Set(UsernameCtx, username)
	c.Set(userRoleCtx, string(userRole))

}
