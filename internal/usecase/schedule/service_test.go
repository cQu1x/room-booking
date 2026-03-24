package schedule_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/schedule"
	"github.com/google/uuid"
)

func TestCreate_Success(t *testing.T) {
	roomID := uuid.New()
	generateCalled := false

	scheduleRepo := &testutil.MockScheduleRepo{
		GetScheduleByRoomIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Schedule, error) {
			return nil, nil // no existing schedule
		},
		CreateScheduleFn: func(_ context.Context, _ *entity.Schedule) error {
			return nil
		},
	}
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Room, error) {
			return &entity.Room{ID: id}, nil
		},
	}
	slotSvc := &testutil.MockSlotService{
		GenerateForScheduleFn: func(_ context.Context, _ *entity.Schedule, _, _ time.Time) error {
			generateCalled = true
			return nil
		},
	}

	svc := schedule.NewService(scheduleRepo, roomRepo, slotSvc)
	sch, err := svc.Create(context.Background(), roomID, []int{1, 2, 3}, "09:00", "17:00")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sch.RoomID != roomID {
		t.Errorf("expected roomID %v, got %v", roomID, sch.RoomID)
	}
	if !generateCalled {
		t.Error("expected GenerateForSchedule to be called")
	}
}

func TestCreate_InvalidDayOfWeek(t *testing.T) {
	svc := schedule.NewService(
		&testutil.MockScheduleRepo{},
		&testutil.MockRoomRepo{},
		&testutil.MockSlotService{},
	)

	_, err := svc.Create(context.Background(), uuid.New(), []int{0}, "09:00", "17:00")
	if err == nil {
		t.Error("expected error for day 0, got nil")
	}

	_, err = svc.Create(context.Background(), uuid.New(), []int{8}, "09:00", "17:00")
	if err == nil {
		t.Error("expected error for day 8, got nil")
	}
}

func TestCreate_RoomNotFound(t *testing.T) {
	scheduleRepo := &testutil.MockScheduleRepo{}
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Room, error) {
			return nil, errors.New("not found")
		},
	}

	svc := schedule.NewService(scheduleRepo, roomRepo, &testutil.MockSlotService{})
	_, err := svc.Create(context.Background(), uuid.New(), []int{1}, "09:00", "17:00")

	if !errors.Is(err, entity.ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestCreate_ScheduleAlreadyExists(t *testing.T) {
	existing := &entity.Schedule{ID: uuid.New(), RoomID: uuid.New()}

	scheduleRepo := &testutil.MockScheduleRepo{
		GetScheduleByRoomIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Schedule, error) {
			return existing, nil
		},
	}
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Room, error) {
			return &entity.Room{ID: id}, nil
		},
	}

	svc := schedule.NewService(scheduleRepo, roomRepo, &testutil.MockSlotService{})
	_, err := svc.Create(context.Background(), uuid.New(), []int{1}, "09:00", "17:00")

	if !errors.Is(err, entity.ErrScheduleExists) {
		t.Errorf("expected ErrScheduleExists, got %v", err)
	}
}

func TestCreate_SlotGenerationFailurePropagates(t *testing.T) {
	generationErr := errors.New("slot generation failed")

	scheduleRepo := &testutil.MockScheduleRepo{
		GetScheduleByRoomIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Schedule, error) {
			return nil, nil
		},
		CreateScheduleFn: func(_ context.Context, _ *entity.Schedule) error {
			return nil
		},
	}
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Room, error) {
			return &entity.Room{ID: id}, nil
		},
	}
	slotSvc := &testutil.MockSlotService{
		GenerateForScheduleFn: func(_ context.Context, _ *entity.Schedule, _, _ time.Time) error {
			return generationErr
		},
	}

	svc := schedule.NewService(scheduleRepo, roomRepo, slotSvc)
	_, err := svc.Create(context.Background(), uuid.New(), []int{1}, "09:00", "17:00")

	if !errors.Is(err, generationErr) {
		t.Errorf("expected slot generation error, got %v", err)
	}
}
