package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type BookingService interface {
	Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*entity.Booking, error)
	Cancel(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*entity.Booking, error)
	ListAll(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	ListMy(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error)
}
