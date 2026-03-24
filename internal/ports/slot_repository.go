package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type SlotRepository interface {
	CreateSlots(ctx context.Context, slots []entity.Slot) error
	GetSlotByID(ctx context.Context, id uuid.UUID) (*entity.Slot, error)
	GetMaxSlotDate(ctx context.Context, roomID uuid.UUID) (*time.Time, error)
	ListAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
}
