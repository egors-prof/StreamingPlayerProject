package driven

import (
	"context"

	"github.com/egors-prof/auth_service/internal/domain"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user domain.User) (err error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
}
