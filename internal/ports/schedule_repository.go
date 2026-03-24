package ports

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type ScheduleRepository interface {
	CreateSchedule(ctx context.Context, schedule *entity.Schedule) error
	GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error)
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
}
