// Package testutil provides test doubles for unit testing use cases.
package testutil

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

// MockRoomRepo is a configurable test double for ports.RoomRepository.
// Set the function fields you need; unset ones panic if called.
type MockRoomRepo struct {
	CreateRoomFn  func(ctx context.Context, room *entity.Room) (*entity.Room, error)
	GetRoomByIDFn func(ctx context.Context, id uuid.UUID) (*entity.Room, error)
	ListRoomsFn   func(ctx context.Context) ([]entity.Room, error)
}

func (m *MockRoomRepo) CreateRoom(ctx context.Context, room *entity.Room) (*entity.Room, error) {
	return m.CreateRoomFn(ctx, room)
}

func (m *MockRoomRepo) GetRoomByID(ctx context.Context, id uuid.UUID) (*entity.Room, error) {
	return m.GetRoomByIDFn(ctx, id)
}

func (m *MockRoomRepo) ListRooms(ctx context.Context) ([]entity.Room, error) {
	return m.ListRoomsFn(ctx)
}

// MockScheduleRepo is a configurable test double for ports.ScheduleRepository.
type MockScheduleRepo struct {
	CreateScheduleFn      func(ctx context.Context, schedule *entity.Schedule) error
	GetScheduleByRoomIDFn func(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error)
}

func (m *MockScheduleRepo) CreateSchedule(ctx context.Context, schedule *entity.Schedule) error {
	return m.CreateScheduleFn(ctx, schedule)
}

func (m *MockScheduleRepo) GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error) {
	return m.GetScheduleByRoomIDFn(ctx, roomID)
}

// MockSlotRepo is a configurable test double for ports.SlotRepository.
type MockSlotRepo struct {
	CreateSlotsFn        func(ctx context.Context, slots []entity.Slot) error
	GetSlotByIDFn        func(ctx context.Context, id uuid.UUID) (*entity.Slot, error)
	GetMaxSlotDateFn     func(ctx context.Context, roomID uuid.UUID) (*time.Time, error)
	ListAvailableSlotsFn func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
}

func (m *MockSlotRepo) CreateSlots(ctx context.Context, slots []entity.Slot) error {
	return m.CreateSlotsFn(ctx, slots)
}

func (m *MockSlotRepo) GetSlotByID(ctx context.Context, id uuid.UUID) (*entity.Slot, error) {
	return m.GetSlotByIDFn(ctx, id)
}

func (m *MockSlotRepo) GetMaxSlotDate(ctx context.Context, roomID uuid.UUID) (*time.Time, error) {
	return m.GetMaxSlotDateFn(ctx, roomID)
}

func (m *MockSlotRepo) ListAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error) {
	return m.ListAvailableSlotsFn(ctx, roomID, date)
}

// MockBookingRepo is a configurable test double for ports.BookingRepository.
type MockBookingRepo struct {
	CreateBookingFn   func(ctx context.Context, booking *entity.Booking) error
	GetBookingByIDFn  func(ctx context.Context, id uuid.UUID) (*entity.Booking, error)
	CancelBookingFn   func(ctx context.Context, id uuid.UUID) error
	ListAllBookingsFn func(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error)
	ListByUserIDFn    func(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error)
	IsSlotBookedFn    func(ctx context.Context, slotID uuid.UUID) (bool, error)
}

func (m *MockBookingRepo) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	return m.CreateBookingFn(ctx, booking)
}

func (m *MockBookingRepo) GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error) {
	return m.GetBookingByIDFn(ctx, id)
}

func (m *MockBookingRepo) CancelBooking(ctx context.Context, id uuid.UUID) error {
	return m.CancelBookingFn(ctx, id)
}

func (m *MockBookingRepo) ListAllBookings(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error) {
	return m.ListAllBookingsFn(ctx, page, pageSize)
}

func (m *MockBookingRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error) {
	return m.ListByUserIDFn(ctx, userID)
}

func (m *MockBookingRepo) IsSlotBooked(ctx context.Context, slotID uuid.UUID) (bool, error) {
	return m.IsSlotBookedFn(ctx, slotID)
}

// MockUserRepo is a configurable test double for ports.UserRepository.
type MockUserRepo struct {
	CreateUserFn     func(ctx context.Context, user *entity.User) (entity.User, error)
	GetUserByEmailFn func(ctx context.Context, email string) (*entity.User, error)
	GetUserByIDFn    func(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *entity.User) (entity.User, error) {
	return m.CreateUserFn(ctx, user)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.GetUserByEmailFn(ctx, email)
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.GetUserByIDFn(ctx, id)
}

// MockSlotService is a configurable test double for ports.SlotService.
type MockSlotService struct {
	ListAvailableFn      func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error)
	GenerateForScheduleFn func(ctx context.Context, schedule *entity.Schedule, from, to time.Time) error
}

func (m *MockSlotService) ListAvailable(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error) {
	return m.ListAvailableFn(ctx, roomID, date)
}

func (m *MockSlotService) GenerateForSchedule(ctx context.Context, schedule *entity.Schedule, from, to time.Time) error {
	return m.GenerateForScheduleFn(ctx, schedule, from, to)
}
