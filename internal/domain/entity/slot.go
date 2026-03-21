package entity

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID        uuid.UUID
	RoomID    int
	StartTime time.Time
	EndTime   time.Time
}

var slotNamespace = uuid.MustParse("31f242fe-a656-4468-99fa-4f91b02283a2")

func GenerateSlotID(roomID int, startTime time.Time) uuid.UUID {
	name := fmt.Sprintf("%s-%s", strconv.Itoa(roomID), startTime.Format(time.RFC3339))
	return uuid.NewSHA1(slotNamespace, []byte(name))
}
