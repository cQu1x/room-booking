package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, room *entity.Room) (*entity.Room, error)
	GetRoomByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
	ListRooms(ctx context.Context) ([]entity.Room, error)
}
