package slot

import (
	"context"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

const (
	slotDuration = 30 * time.Minute
	windowDays   = 7
)

type Service struct {
	slotRepo     ports.SlotRepository
	scheduleRepo ports.ScheduleRepository
}

func NewService(slotRepo ports.SlotRepository, scheduleRepo ports.ScheduleRepository) *Service {
	return &Service{slotRepo: slotRepo, scheduleRepo: scheduleRepo}
}

// ListAvailable returns slots for the room on the given date that have no active booking.
// It also fires a non-blocking goroutine to extend the slot window if it is about to run out.
func (s *Service) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error) {
	slots, err := s.slotRepo.ListAvailableSlots(ctx, roomID, date)
	if err != nil {
		return nil, err
	}
	// Fire-and-forget: extend the window in the background so the response is never blocked.
	go s.ensureWindow(context.Background(), roomID)
	return slots, nil
}

// ensureWindow checks whether the slot horizon for the room is still at least
// windowDays ahead. If not, it generates the missing days. It is safe to call
// concurrently — duplicate slots are silently ignored (ON CONFLICT DO NOTHING).
func (s *Service) ensureWindow(ctx context.Context, roomID uuid.UUID) {
	maxDate, err := s.slotRepo.GetMaxSlotDate(ctx, roomID)
	if err != nil {
		return
	}

	now := time.Now().UTC()
	horizon := now.AddDate(0, 0, windowDays)

	// Nothing to do if the window already extends far enough.
	if maxDate != nil && !maxDate.Before(horizon) {
		return
	}

	schedule, err := s.scheduleRepo.GetScheduleByRoomID(ctx, roomID)
	if err != nil || schedule == nil {
		return // room has no schedule; slots are never generated for it
	}

	// Generate from the day after the current max (or today if no slots exist yet).
	var from time.Time
	if maxDate == nil {
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		next := maxDate.AddDate(0, 0, 1)
		from = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
	}
	to := time.Date(horizon.Year(), horizon.Month(), horizon.Day()+1, 0, 0, 0, 0, time.UTC)

	_ = s.GenerateForSchedule(ctx, schedule, from, to)
}

// GenerateForSchedule creates all 30-minute slots for the schedule in [from, to).
// Slots are inserted with ON CONFLICT DO NOTHING so calling this multiple times is safe.
func (s *Service) GenerateForSchedule(ctx context.Context, schedule *entity.Schedule, from, to time.Time) error {
	startHour, startMinute, err := parseHHMM(schedule.StartTime)
	if err != nil {
		return err
	}
	endHour, endMinute, err := parseHHMM(schedule.EndTime)
	if err != nil {
		return err
	}

	daySet := make(map[time.Weekday]bool, len(schedule.DaysOfWeek))
	for _, dow := range schedule.DaysOfWeek {
		daySet[isoToWeekday(dow)] = true
	}

	var slots []entity.Slot
	for date := from; date.Before(to); date = date.AddDate(0, 0, 1) {
		if !daySet[date.Weekday()] {
			continue
		}
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMinute, 0, 0, time.UTC)
		dayEnd := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMinute, 0, 0, time.UTC)

		for slotStart := dayStart; !slotStart.Add(slotDuration).After(dayEnd); slotStart = slotStart.Add(slotDuration) {
			slots = append(slots, entity.Slot{
				ID:        entity.GenerateSlotID(schedule.RoomID, slotStart),
				RoomID:    schedule.RoomID,
				StartTime: slotStart,
				EndTime:   slotStart.Add(slotDuration),
			})
		}
	}

	return s.slotRepo.CreateSlots(ctx, slots)
}

func parseHHMM(timeStr string) (int, int, error) {
	var hour, minute int
	if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute); err != nil {
		return 0, 0, fmt.Errorf("invalid time format %q: %w", timeStr, err)
	}
	return hour, minute, nil
}

// isoToWeekday converts ISO weekday (1=Mon … 7=Sun) to time.Weekday.
func isoToWeekday(isoDay int) time.Weekday {
	if isoDay == 7 {
		return time.Sunday
	}
	return time.Weekday(isoDay) // 1→Mon, 2→Tue, …, 6→Sat
}
