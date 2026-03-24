package slot_test

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/slot"
	"github.com/google/uuid"
)

func TestGenerateForSchedule_CreatesCorrectSlots(t *testing.T) {
	roomID := uuid.New()

	schedule := &entity.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	from := nextWeekday(time.Monday)
	to := from.AddDate(0, 0, 1)
	var created []entity.Slot
	slotRepo := &testutil.MockSlotRepo{
		CreateSlotsFn: func(_ context.Context, slots []entity.Slot) error {
			created = slots
			return nil
		},
	}

	svc := slot.NewService(slotRepo, &testutil.MockScheduleRepo{})
	err := svc.GenerateForSchedule(context.Background(), schedule, from, to)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(created) != 2 {
		t.Errorf("expected 2 slots, got %d", len(created))
	}
	for _, s := range created {
		if s.RoomID != roomID {
			t.Errorf("slot has wrong roomID: %v", s.RoomID)
		}
		if s.EndTime.Sub(s.StartTime) != 30*time.Minute {
			t.Errorf("slot duration should be 30 minutes, got %v", s.EndTime.Sub(s.StartTime))
		}
	}
}

func TestGenerateForSchedule_SkipsNonScheduledDays(t *testing.T) {

	monday := nextWeekday(time.Monday)
	tuesday := monday.AddDate(0, 0, 1)

	schedule := &entity.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	var created []entity.Slot
	slotRepo := &testutil.MockSlotRepo{
		CreateSlotsFn: func(_ context.Context, slots []entity.Slot) error {
			created = slots
			return nil
		},
	}

	svc := slot.NewService(slotRepo, &testutil.MockScheduleRepo{})
	err := svc.GenerateForSchedule(context.Background(), schedule, monday, tuesday.AddDate(0, 0, 1))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(created) != 2 {
		t.Errorf("expected 2 slots for Monday only, got %d", len(created))
	}
}

func TestGenerateForSchedule_DeterministicIDs(t *testing.T) {

	monday := nextWeekday(time.Monday)
	schedule := &entity.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "09:30",
	}

	var firstRun, secondRun []entity.Slot
	slotRepo := &testutil.MockSlotRepo{
		CreateSlotsFn: func(_ context.Context, slots []entity.Slot) error {
			if firstRun == nil {
				firstRun = slots
			} else {
				secondRun = slots
			}
			return nil
		},
	}

	svc := slot.NewService(slotRepo, &testutil.MockScheduleRepo{})
	to := monday.AddDate(0, 0, 1)
	_ = svc.GenerateForSchedule(context.Background(), schedule, monday, to)
	_ = svc.GenerateForSchedule(context.Background(), schedule, monday, to)

	if len(firstRun) != len(secondRun) {
		t.Fatalf("slot counts differ: %d vs %d", len(firstRun), len(secondRun))
	}
	for i := range firstRun {
		if firstRun[i].ID != secondRun[i].ID {
			t.Errorf("slot ID mismatch at index %d: %v != %v", i, firstRun[i].ID, secondRun[i].ID)
		}
	}
}

func TestGenerateForSchedule_EmptyRangeProducesNoSlots(t *testing.T) {
	now := time.Now().UTC()
	schedule := &entity.Schedule{
		ID:         uuid.New(),
		RoomID:     uuid.New(),
		DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}

	var created []entity.Slot
	slotRepo := &testutil.MockSlotRepo{
		CreateSlotsFn: func(_ context.Context, slots []entity.Slot) error {
			created = slots
			return nil
		},
	}

	svc := slot.NewService(slotRepo, &testutil.MockScheduleRepo{})

	err := svc.GenerateForSchedule(context.Background(), schedule, now, now)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(created) != 0 {
		t.Errorf("expected no slots for empty range, got %d", len(created))
	}
}

func TestListAvailable_ReturnsRepositoryResult(t *testing.T) {
	roomID := uuid.New()
	date := time.Now().UTC()
	expected := []entity.Slot{
		{ID: uuid.New(), RoomID: roomID, StartTime: date, EndTime: date.Add(30 * time.Minute)},
	}

	slotRepo := &testutil.MockSlotRepo{
		ListAvailableSlotsFn: func(_ context.Context, rid uuid.UUID, d time.Time) ([]entity.Slot, error) {
			return expected, nil
		},
		GetMaxSlotDateFn: func(_ context.Context, _ uuid.UUID) (*time.Time, error) {
			t := time.Now().UTC().AddDate(1, 0, 0)
			return &t, nil
		},
	}
	scheduleRepo := &testutil.MockScheduleRepo{
		GetScheduleByRoomIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Schedule, error) {
			return nil, nil
		},
	}

	svc := slot.NewService(slotRepo, scheduleRepo)
	slots, err := svc.ListAvailable(context.Background(), roomID, date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != len(expected) {
		t.Errorf("expected %d slots, got %d", len(expected), len(slots))
	}
}

func nextWeekday(wd time.Weekday) time.Time {
	now := time.Now().UTC()
	d := now
	for d.Weekday() != wd {
		d = d.AddDate(0, 0, 1)
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}
