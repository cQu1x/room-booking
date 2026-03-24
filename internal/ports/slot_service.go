package ports

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type SlotService interface {
	ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
	GenerateForSchedule(ctx context.Context, schedule *entity.Schedule, from, to time.Time) error
}
