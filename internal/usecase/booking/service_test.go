package booking_test

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/booking"
	"github.com/google/uuid"
)

func TestCreate_Success(t *testing.T) {
	slotID := uuid.New()
	userID := uuid.New()
	futureSlot := &entity.Slot{
		ID:        slotID,
		RoomID:    uuid.New(),
		StartTime: time.Now().UTC().Add(time.Hour),
		EndTime:   time.Now().UTC().Add(90 * time.Minute),
	}

	slotRepo := &testutil.MockSlotRepo{
		GetSlotByIDFn: func(_ context.Context, id uuid.UUID) (*entity.Slot, error) {
			return futureSlot, nil
		},
	}
	bookingRepo := &testutil.MockBookingRepo{
		IsSlotBookedFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return false, nil
		},
		CreateBookingFn: func(_ context.Context, _ *entity.Booking) error {
			return nil
		},
	}

	svc := booking.NewService(bookingRepo, slotRepo)
	b, err := svc.Create(context.Background(), userID, slotID, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, b.UserID)
	}
	if b.SlotID != slotID {
		t.Errorf("expected slotID %v, got %v", slotID, b.SlotID)
	}
	if b.Status != entity.BookingStatusActive {
		t.Errorf("expected status active, got %v", b.Status)
	}
	if b.ConferenceLink != nil {
		t.Error("expected no conference link")
	}
}

func TestCreate_WithConferenceLink(t *testing.T) {
	slotID := uuid.New()
	slotRepo := &testutil.MockSlotRepo{
		GetSlotByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Slot, error) {
			return &entity.Slot{
				ID:        slotID,
				StartTime: time.Now().UTC().Add(time.Hour),
				EndTime:   time.Now().UTC().Add(90 * time.Minute),
			}, nil
		},
	}
	bookingRepo := &testutil.MockBookingRepo{
		IsSlotBookedFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return false, nil
		},
		CreateBookingFn: func(_ context.Context, _ *entity.Booking) error {
			return nil
		},
	}

	svc := booking.NewService(bookingRepo, slotRepo)
	b, err := svc.Create(context.Background(), uuid.New(), slotID, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ConferenceLink == nil {
		t.Error("expected conference link to be set")
	}
}

func TestCreate_SlotNotFound(t *testing.T) {
	slotRepo := &testutil.MockSlotRepo{
		GetSlotByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Slot, error) {
			return nil, entity.ErrSlotNotFound
		},
	}

	svc := booking.NewService(&testutil.MockBookingRepo{}, slotRepo)
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)

	if err != entity.ErrSlotNotFound {
		t.Errorf("expected ErrSlotNotFound, got %v", err)
	}
}

func TestCreate_SlotInPast(t *testing.T) {
	slotRepo := &testutil.MockSlotRepo{
		GetSlotByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Slot, error) {
			return &entity.Slot{
				StartTime: time.Now().UTC().Add(-time.Hour), // past
			}, nil
		},
	}

	svc := booking.NewService(&testutil.MockBookingRepo{}, slotRepo)
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)

	if err != entity.ErrSlotInPast {
		t.Errorf("expected ErrSlotInPast, got %v", err)
	}
}

func TestCreate_SlotAlreadyBooked(t *testing.T) {
	slotRepo := &testutil.MockSlotRepo{
		GetSlotByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Slot, error) {
			return &entity.Slot{
				StartTime: time.Now().UTC().Add(time.Hour),
			}, nil
		},
	}
	bookingRepo := &testutil.MockBookingRepo{
		IsSlotBookedFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	svc := booking.NewService(bookingRepo, slotRepo)
	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), false)

	if err != entity.ErrSlotAlreadyBooked {
		t.Errorf("expected ErrSlotAlreadyBooked, got %v", err)
	}
}

func TestCancel_Success(t *testing.T) {
	bookingID := uuid.New()
	userID := uuid.New()
	cancelCalled := false

	bookingRepo := &testutil.MockBookingRepo{
		GetBookingByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Booking, error) {
			return &entity.Booking{
				ID:     bookingID,
				UserID: userID,
				Status: entity.BookingStatusActive,
			}, nil
		},
		CancelBookingFn: func(_ context.Context, _ uuid.UUID) error {
			cancelCalled = true
			return nil
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	b, err := svc.Cancel(context.Background(), userID, bookingID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != entity.BookingStatusCancelled {
		t.Errorf("expected status cancelled, got %v", b.Status)
	}
	if !cancelCalled {
		t.Error("expected CancelBooking to be called")
	}
}

func TestCancel_Idempotent(t *testing.T) {
	// Cancelling an already-cancelled booking returns 200 without calling CancelBooking again.
	bookingID := uuid.New()
	userID := uuid.New()
	cancelCalled := false

	bookingRepo := &testutil.MockBookingRepo{
		GetBookingByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Booking, error) {
			return &entity.Booking{
				ID:     bookingID,
				UserID: userID,
				Status: entity.BookingStatusCancelled,
			}, nil
		},
		CancelBookingFn: func(_ context.Context, _ uuid.UUID) error {
			cancelCalled = true
			return nil
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	b, err := svc.Cancel(context.Background(), userID, bookingID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Status != entity.BookingStatusCancelled {
		t.Errorf("expected status cancelled, got %v", b.Status)
	}
	if cancelCalled {
		t.Error("CancelBooking must not be called for an already-cancelled booking")
	}
}

func TestCancel_NotFound(t *testing.T) {
	bookingRepo := &testutil.MockBookingRepo{
		GetBookingByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Booking, error) {
			return nil, entity.ErrBookingNotFound
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	_, err := svc.Cancel(context.Background(), uuid.New(), uuid.New())

	if err != entity.ErrBookingNotFound {
		t.Errorf("expected ErrBookingNotFound, got %v", err)
	}
}

func TestCancel_Forbidden(t *testing.T) {
	bookingRepo := &testutil.MockBookingRepo{
		GetBookingByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Booking, error) {
			return &entity.Booking{
				ID:     uuid.New(),
				UserID: uuid.New(), // different owner
				Status: entity.BookingStatusActive,
			}, nil
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	_, err := svc.Cancel(context.Background(), uuid.New(), uuid.New())

	if err != entity.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestListAll(t *testing.T) {
	page, pageSize := 1, 10
	expected := []entity.Booking{{ID: uuid.New()}, {ID: uuid.New()}}

	bookingRepo := &testutil.MockBookingRepo{
		ListAllBookingsFn: func(_ context.Context, p, ps int) ([]entity.Booking, int, error) {
			if p != page || ps != pageSize {
				t.Errorf("unexpected pagination: page=%d pageSize=%d", p, ps)
			}
			return expected, len(expected), nil
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	bookings, total, err := svc.ListAll(context.Background(), page, pageSize)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != len(expected) {
		t.Errorf("expected %d bookings, got %d", len(expected), len(bookings))
	}
	if total != len(expected) {
		t.Errorf("expected total %d, got %d", len(expected), total)
	}
}

func TestListMy(t *testing.T) {
	userID := uuid.New()
	expected := []entity.Booking{{ID: uuid.New(), UserID: userID}}

	bookingRepo := &testutil.MockBookingRepo{
		ListByUserIDFn: func(_ context.Context, id uuid.UUID) ([]entity.Booking, error) {
			if id != userID {
				t.Errorf("unexpected userID: %v", id)
			}
			return expected, nil
		},
	}

	svc := booking.NewService(bookingRepo, &testutil.MockSlotRepo{})
	bookings, err := svc.ListMy(context.Background(), userID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}
}
