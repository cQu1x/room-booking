package entity

import "errors"

var ErrUserNotFound = errors.New("user not found")

type DomainError struct {
	msg        string
	httpStatus int
	httpCode   string
}

func (e *DomainError) Error() string    { return e.msg }
func (e *DomainError) HTTPStatus() int  { return e.httpStatus }
func (e *DomainError) HTTPCode() string { return e.httpCode }

var (
	ErrRoomNotFound       = &DomainError{"room not found", 404, "ROOM_NOT_FOUND"}
	ErrScheduleExists     = &DomainError{"schedule already exists for this room", 409, "SCHEDULE_EXISTS"}
	ErrSlotNotFound       = &DomainError{"slot not found", 404, "SLOT_NOT_FOUND"}
	ErrSlotAlreadyBooked  = &DomainError{"slot is already booked", 409, "SLOT_ALREADY_BOOKED"}
	ErrSlotInPast         = &DomainError{"cannot book a slot in the past", 400, "INVALID_REQUEST"}
	ErrBookingNotFound    = &DomainError{"booking not found", 404, "BOOKING_NOT_FOUND"}
	ErrForbidden          = &DomainError{"forbidden", 403, "FORBIDDEN"}
	ErrEmailAlreadyTaken  = &DomainError{"email already taken", 400, "INVALID_REQUEST"}
	ErrInvalidCredentials = &DomainError{"invalid email or password", 401, "UNAUTHORIZED"}
)
