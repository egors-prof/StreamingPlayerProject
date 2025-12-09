package app

import (
	"fmt"

	"github.com/egors-prof/searchService/internal/config"
	Server "github.com/egors-prof/searchService/internal/delivery/http"
)


type App struct{
	Server *Server.Server
}
func(a*App)Run(cfg config.Config){
	a.Server.RegisterRoutes()
	a.Server.Router.Run(fmt.Sprintf(":%s",cfg.HTTPPort))
	
}