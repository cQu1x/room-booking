package entity

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	SlotID    uuid.UUID
	Status    BookingStatus
	createdAt time.Time
}
