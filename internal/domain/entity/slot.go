package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID        uuid.UUID
	RoomID    uuid.UUID
	StartTime time.Time
	EndTime   time.Time
}

var slotNamespace = uuid.MustParse("31f242fe-a656-4468-99fa-4f91b02283a2")

// GenerateSlotID возвращает детерминированный UUID слота на основе переговорки и времени начала.
func GenerateSlotID(roomID uuid.UUID, startTime time.Time) uuid.UUID {
	name := fmt.Sprintf("%s-%s", roomID.String(), startTime.UTC().Format(time.RFC3339))
	return uuid.NewSHA1(slotNamespace, []byte(name))
}
