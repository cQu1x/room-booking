package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type ScheduleService interface {
	Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*entity.Schedule, error)
}
