package main

import (
	"log"

	"github.com/egors-prof/searchService/internal/app"
	"github.com/egors-prof/searchService/internal/config"
	Server "github.com/egors-prof/searchService/internal/delivery/http"
	"github.com/egors-prof/searchService/internal/repository"
	"github.com/egors-prof/searchService/internal/usecases"
)

func main(){
	cfg,err:=config.GetConfigs()
	if err!=nil{
		log.Fatalln(err)
	}
	

	r,err:=repository.InitRepository(&cfg)
	if err!=nil{
		log.Fatalln(err)
	}
	
	usecases:=usecases.NewUseCases(r)

	http:=Server.NewServer(usecases)
	app:=&app.App{Server:http}
	app.Run(cfg)

}