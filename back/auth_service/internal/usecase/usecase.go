package usecase

import (
	"github.com/egors-prof/auth_service/internal/adapter/driven/broker"
	"github.com/egors-prof/auth_service/internal/adapter/driven/dbstore"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/auth_service/internal/port/usecase"
	authenticate "github.com/egors-prof/auth_service/internal/usecase/authenticator"
	usercreater "github.com/egors-prof/auth_service/internal/usecase/user_creater"
)

type UseCases struct {
	UserCreater   usecase.UserCreater
	Authenticator usecase.Authenticate
}

func New(cfg config.Config, store *dbstore.DBStore, publisher *broker.MessagePublisher) *UseCases {
	return &UseCases{
		UserCreater:   usercreater.New(&cfg, store.UserStorage, publisher),
		Authenticator: authenticate.New(&cfg, store.UserStorage),
	}
}
