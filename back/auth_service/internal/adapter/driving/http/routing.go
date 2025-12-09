package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) endpoints() {
	s.router.GET("/ping", s.Ping)
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authG := s.router.Group("/auth")
	authG.POST("/sign-up", s.SignUp)
	authG.POST("/sign-in", s.SignIn)
	authG.GET("/refresh", s.RefreshTokenPair)
}

func (s *Server) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ping": "pong",
	})
}
