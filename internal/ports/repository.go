package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *entity.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
	List(ctx context.Context) ([]entity.Room, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *entity.Schedule) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error)
}

type SlotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Slot, error)
	ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking *entity.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	IsSlotBooked(ctx context.Context, slotID uuid.UUID) (bool, error)
}

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
