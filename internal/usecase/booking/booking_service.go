package booking

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	bookingRepo ports.BookingRepository
	slotRepo    ports.SlotRepository
}

func NewService(bookingRepo ports.BookingRepository, slotRepo ports.SlotRepository) *Service {
	return &Service{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*entity.Booking, error) {
	slot, err := s.slotRepo.GetSlotByID(ctx, slotID)
	if err != nil {
		return nil, entity.ErrSlotNotFound
	}

	if slot.StartTime.Before(time.Now().UTC()) {
		return nil, entity.ErrSlotInPast
	}

	booked, err := s.bookingRepo.IsSlotBooked(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if booked {
		return nil, entity.ErrSlotAlreadyBooked
	}

	booking := &entity.Booking{
		ID:        uuid.New(),
		UserID:    userID,
		SlotID:    slotID,
		Status:    entity.BookingStatusActive,
		CreatedAt: time.Now().UTC(),
	}

	if createConferenceLink {
		link := "https://conference.example.com/meeting/" + uuid.New().String()
		booking.ConferenceLink = &link
	}

	if err := s.bookingRepo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *Service) Cancel(ctx context.Context, userID uuid.UUID, bookingID uuid.UUID) (*entity.Booking, error) {
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, entity.ErrBookingNotFound
	}

	if booking.UserID != userID {
		return nil, entity.ErrForbidden
	}

	if booking.Status == entity.BookingStatusCancelled {
		return booking, nil
	}

	if err := s.bookingRepo.CancelBooking(ctx, bookingID); err != nil {
		return nil, err
	}

	booking.Status = entity.BookingStatusCancelled
	return booking, nil
}

func (s *Service) ListAll(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error) {
	return s.bookingRepo.ListAllBookings(ctx, page, pageSize)
}

func (s *Service) ListMy(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error) {
	return s.bookingRepo.ListByUserID(ctx, userID)
}
