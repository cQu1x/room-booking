package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type RoomService interface {
	Create(ctx context.Context, name string, description *string, capacity *int) (*entity.Room, error)
	List(ctx context.Context) ([]entity.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
}
