package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type AuthService interface {
	DummyLogin(ctx context.Context, role entity.Role) (string, error)
	Register(ctx context.Context, email, password string, role entity.Role) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type RoomService interface {
	Create(ctx context.Context, name string, description *string, capacity *int) (*entity.Room, error)
	List(ctx context.Context) ([]entity.Room, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
}

type ScheduleService interface {
	Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*entity.Schedule, error)
}

type SlotService interface {
	ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
	GenerateForSchedule(ctx context.Context, schedule *entity.Schedule, from, to time.Time) error
}

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*entity.Booking, error)
	Cancel(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*entity.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	ListMy(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error)
}
