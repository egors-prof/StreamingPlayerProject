package usercreater

import (
	"context"
	"errors"
	"fmt"

	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/auth_service/internal/domain"
	"github.com/egors-prof/auth_service/internal/errs"
	"github.com/egors-prof/auth_service/internal/port/driven"
	"github.com/egors-prof/auth_service/internal/utils"
)

type UseCase struct {
	cfg              *config.Config
	userStorage      driven.UserStorage
	messagePublisher driven.MessagePublisher
}

func New(cfg *config.Config, userStorage driven.UserStorage, publisher driven.MessagePublisher) *UseCase {
	return &UseCase{
		cfg:              cfg,
		userStorage:      userStorage,
		messagePublisher: publisher,
	}
}

func (u *UseCase) CreateUser(ctx context.Context, user domain.User) (err error) {
	// Проверить существует ли пользователь с таким username'ом в бд
	_, err = u.userStorage.GetUserByUsername(ctx, user.Username)
	if err != nil {
		if !errors.Is(err, errs.ErrNotfound) {
			return err
		}
	} else {
		return errs.ErrUsernameAlreadyExists
	}

	// За хэшировать пароль
	user.Password, err = utils.GenerateHash(user.Password)
	if err != nil {
		return err
	}

	user.Role = domain.RoleUser

	// Добавить пользователя в бд
	if err = u.userStorage.CreateUser(ctx, user); err != nil {
		return err
	}

	// Publish message to message broker
	if u.messagePublisher != nil {
		message := domain.Message{
			Recipient: user.Username,
			Subject:   fmt.Sprintf("Welcome, %s!", user.FullName),
			Body:      fmt.Sprintf("Hello, %s! Your account has been created successfully.", user.FullName),
		}
		if err = u.messagePublisher.PublishMessage(message); err != nil {
			// Just log the error as a demo behavior
			fmt.Println(err)
		}
	}

	return nil
}
