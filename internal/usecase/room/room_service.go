package room

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	roomRepo ports.RoomRepository
}

func NewService(roomRepo ports.RoomRepository) *Service {
	return &Service{roomRepo: roomRepo}
}

func (s *Service) Create(ctx context.Context, name string, description *string, capacity *int) (*entity.Room, error) {
	room := &entity.Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC(),
	}
	return s.roomRepo.CreateRoom(ctx, room)
}

func (s *Service) List(ctx context.Context) ([]entity.Room, error) {
	return s.roomRepo.ListRooms(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error) {
	return s.roomRepo.GetRoomByID(ctx, id)
}
