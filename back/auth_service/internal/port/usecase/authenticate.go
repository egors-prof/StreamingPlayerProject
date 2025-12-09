package usecase

import (
	"context"

	"github.com/egors-prof/auth_service/internal/domain"
)

type Authenticate interface {
	Authenticate(ctx context.Context, user domain.User) (int, domain.Role, error)
}
