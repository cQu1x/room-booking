package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) error
	GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	CancelBooking(ctx context.Context, id uuid.UUID) error
	ListAllBookings(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error)
	IsSlotBooked(ctx context.Context, slotID uuid.UUID) (bool, error)
}
