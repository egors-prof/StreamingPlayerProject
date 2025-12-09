package usecase

import (
	"context"

	"github.com/egors-prof/auth_service/internal/domain"
)

type UserCreater interface {
	CreateUser(ctx context.Context, user domain.User) (err error)
}
