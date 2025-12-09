package Server

import (
	"net/http"

	"github.com/egors-prof/searchService/internal/usecases"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct{
	Router *gin.Engine
	Uc *usecases.UseCases
}

func NewServer(uc *usecases.UseCases)*Server{
	
	r := gin.Default()
	
	// CORS middleware
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"}, // Allow all in dev
        AllowMethods:     []string{"GET", "POST", "OPTIONS"},
        AllowHeaders:     []string{"*"}, // Allow all headers
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))
    
    // Or simpler CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
            return
        }
        
        c.Next()
    })
	r.Use(cors.Default())
	return &Server{
		Router: r,
		Uc:uc,
	}
}