package entity

import "errors"

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrScheduleExists     = errors.New("schedule already exists for this room")
	ErrSlotNotFound       = errors.New("slot not found")
	ErrSlotAlreadyBooked  = errors.New("slot is already booked")
	ErrSlotInPast         = errors.New("cannot book a slot in the past")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrForbidden          = errors.New("forbidden")
	ErrEmailAlreadyTaken  = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
)
