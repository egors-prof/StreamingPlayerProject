package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/egors-prof/streaming/internal/config"
	"github.com/egors-prof/streaming/internal/usecase"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
	uc     *usecase.UseCases
}

const httpServerReadHeaderTimeout = 70 * time.Second

func New(
	cfg *config.Config,
	uc *usecase.UseCases,
) *http.Server {
	r := gin.New()
	
	srv := &Server{
		router: r,
		cfg:    cfg,
		uc:     uc,
	}

	srv.endpoints()

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           srv,
		ReadHeaderTimeout: httpServerReadHeaderTimeout,
	}

	return httpServer
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
