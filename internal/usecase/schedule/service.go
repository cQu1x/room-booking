package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

const slotGenerationDays = 7

type Service struct {
	scheduleRepo ports.ScheduleRepository
	roomRepo     ports.RoomRepository
	slotService  ports.SlotService
}

// NewService создаёт сервис расписаний.
func NewService(scheduleRepo ports.ScheduleRepository, roomRepo ports.RoomRepository, slotService ports.SlotService) *Service {
	return &Service{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		slotService:  slotService,
	}
}

// Create создаёт расписание для переговорки и предгенерирует слоты на ближайшие 7 дней.
func (s *Service) Create(ctx context.Context, roomID uuid.UUID, daysOfWeek []int, startTime, endTime string) (*entity.Schedule, error) {
	for _, d := range daysOfWeek {
		if d < 1 || d > 7 {
			return nil, fmt.Errorf("invalid day of week %d: must be in range 1-7", d)
		}
	}

	if _, err := s.roomRepo.GetRoomByID(ctx, roomID); err != nil {
		return nil, entity.ErrRoomNotFound
	}

	existing, err := s.scheduleRepo.GetScheduleByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, entity.ErrScheduleExists
	}

	schedule := &entity.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	if err := s.scheduleRepo.CreateSchedule(ctx, schedule); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, slotGenerationDays)
	if err := s.slotService.GenerateForSchedule(ctx, schedule, from, to); err != nil {
		return nil, err
	}

	return schedule, nil
}
