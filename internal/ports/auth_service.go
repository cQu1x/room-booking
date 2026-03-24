package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
)

type AuthService interface {
	DummyLogin(ctx context.Context, role entity.Role) (string, error)
	Register(ctx context.Context, email, password string, role entity.Role) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}
