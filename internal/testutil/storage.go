package testutil

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

type InMemStore struct {
	mu           sync.RWMutex
	rooms        map[uuid.UUID]*entity.Room
	schedules    map[uuid.UUID]*entity.Schedule
	slots        map[uuid.UUID]*entity.Slot
	bookings     map[uuid.UUID]*entity.Booking
	users        map[uuid.UUID]*entity.User
	usersByEmail map[string]*entity.User
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		rooms:        make(map[uuid.UUID]*entity.Room),
		schedules:    make(map[uuid.UUID]*entity.Schedule),
		slots:        make(map[uuid.UUID]*entity.Slot),
		bookings:     make(map[uuid.UUID]*entity.Booking),
		users:        make(map[uuid.UUID]*entity.User),
		usersByEmail: make(map[string]*entity.User),
	}
}

// ── RoomRepository ────────────────────────────────────────────────────────────

func (s *InMemStore) CreateRoom(_ context.Context, room *entity.Room) (*entity.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[room.ID] = room
	return room, nil
}

func (s *InMemStore) GetRoomByID(_ context.Context, id uuid.UUID) (*entity.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room, ok := s.rooms[id]
	if !ok {
		return nil, entity.ErrRoomNotFound
	}
	return room, nil
}

func (s *InMemStore) ListRooms(_ context.Context) ([]entity.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rooms := make([]entity.Room, 0, len(s.rooms))
	for _, r := range s.rooms {
		rooms = append(rooms, *r)
	}
	return rooms, nil
}

// ── ScheduleRepository ────────────────────────────────────────────────────────

func (s *InMemStore) CreateSchedule(_ context.Context, schedule *entity.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[schedule.RoomID] = schedule
	return nil
}

func (s *InMemStore) GetScheduleByRoomID(_ context.Context, roomID uuid.UUID) (*entity.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sch, ok := s.schedules[roomID]
	if !ok {
		return nil, nil
	}
	return sch, nil
}

func (s *InMemStore) DeleteSchedule(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for roomID, sch := range s.schedules {
		if sch.ID == id {
			delete(s.schedules, roomID)
			return nil
		}
	}
	return nil
}

// ── SlotRepository ────────────────────────────────────────────────────────────

func (s *InMemStore) CreateSlots(_ context.Context, slots []entity.Slot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, slot := range slots {
		if _, exists := s.slots[slot.ID]; !exists {
			cp := slot
			s.slots[slot.ID] = &cp
		}
	}
	return nil
}

func (s *InMemStore) GetSlotByID(_ context.Context, id uuid.UUID) (*entity.Slot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	slot, ok := s.slots[id]
	if !ok {
		return nil, entity.ErrSlotNotFound
	}
	return slot, nil
}

func (s *InMemStore) GetMaxSlotDate(_ context.Context, roomID uuid.UUID) (*time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var max *time.Time
	for _, slot := range s.slots {
		if slot.RoomID == roomID {
			if max == nil || slot.StartTime.After(*max) {
				t := slot.StartTime
				max = &t
			}
		}
	}
	return max, nil
}

func (s *InMemStore) ListAvailableSlots(_ context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookedSlots := make(map[uuid.UUID]bool)
	for _, b := range s.bookings {
		if b.Status == entity.BookingStatusActive {
			bookedSlots[b.SlotID] = true
		}
	}

	var available []entity.Slot
	for _, slot := range s.slots {
		if slot.RoomID != roomID {
			continue
		}
		y, m, d := slot.StartTime.Date()
		dy, dm, dd := date.Date()
		if y != dy || m != dm || d != dd {
			continue
		}
		if !bookedSlots[slot.ID] {
			available = append(available, *slot)
		}
	}

	sort.Slice(available, func(i, j int) bool {
		return available[i].StartTime.Before(available[j].StartTime)
	})
	return available, nil
}

// ── BookingRepository ─────────────────────────────────────────────────────────

func (s *InMemStore) CreateBooking(_ context.Context, booking *entity.Booking) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookings[booking.ID] = booking
	return nil
}

func (s *InMemStore) GetBookingByID(_ context.Context, id uuid.UUID) (*entity.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bookings[id]
	if !ok {
		return nil, entity.ErrBookingNotFound
	}
	return b, nil
}

func (s *InMemStore) CancelBooking(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	if !ok {
		return entity.ErrBookingNotFound
	}
	b.Status = entity.BookingStatusCancelled
	return nil
}

func (s *InMemStore) ListAllBookings(_ context.Context, page, pageSize int) ([]entity.Booking, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]entity.Booking, 0, len(s.bookings))
	for _, b := range s.bookings {
		all = append(all, *b)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.Before(all[j].CreatedAt)
	})
	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return []entity.Booking{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func (s *InMemStore) ListByUserID(_ context.Context, userID uuid.UUID) ([]entity.Booking, error) {
	s.mu.RLock()
	now := time.Now().UTC()
	var result []entity.Booking
	startTimes := make(map[uuid.UUID]time.Time)
	for _, b := range s.bookings {
		if b.UserID != userID {
			continue
		}
		slot, ok := s.slots[b.SlotID]
		if !ok || !slot.StartTime.After(now) {
			continue
		}
		result = append(result, *b)
		startTimes[b.SlotID] = slot.StartTime
	}
	s.mu.RUnlock()

	sort.Slice(result, func(i, j int) bool {
		return startTimes[result[i].SlotID].Before(startTimes[result[j].SlotID])
	})
	return result, nil
}

func (s *InMemStore) IsSlotBooked(_ context.Context, slotID uuid.UUID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.bookings {
		if b.SlotID == slotID && b.Status == entity.BookingStatusActive {
			return true, nil
		}
	}
	return false, nil
}

// ── UserRepository ────────────────────────────────────────────────────────────

func (s *InMemStore) CreateUser(_ context.Context, user *entity.User) (entity.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
	s.usersByEmail[user.Email] = user
	return *user, nil
}

func (s *InMemStore) GetUserByEmail(_ context.Context, email string) (*entity.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByEmail[email]
	if !ok {
		return nil, entity.ErrUserNotFound
	}
	return u, nil
}

func (s *InMemStore) GetUserByID(_ context.Context, id uuid.UUID) (*entity.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, entity.ErrUserNotFound
	}
	return u, nil
}
