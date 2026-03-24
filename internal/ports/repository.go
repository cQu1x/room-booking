package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type RoomRepository interface {
	CreateRoom(ctx context.Context, room *entity.Room) (*entity.Room, error)
	GetRoomByID(ctx context.Context, id uuid.UUID) (*entity.Room, error)
	ListRooms(ctx context.Context) ([]entity.Room, error)
}

type ScheduleRepository interface {
	CreateSchedule(ctx context.Context, schedule *entity.Schedule) error
	GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error)
}

type SlotRepository interface {
	CreateSlots(ctx context.Context, slots []entity.Slot) error
	GetSlotByID(ctx context.Context, id uuid.UUID) (*entity.Slot, error)
	GetMaxSlotDate(ctx context.Context, roomID uuid.UUID) (*time.Time, error)
	ListAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
}

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) error
	GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	CancelBooking(ctx context.Context, id uuid.UUID) error
	ListAllBookings(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error)
	IsSlotBooked(ctx context.Context, slotID uuid.UUID) (bool, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) (entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
