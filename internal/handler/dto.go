package handler

import (
	"errors"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
)

// ── Auth ─────────────────────────────────────────────────────────────────────

type dummyLoginRequest struct {
	Role string `json:"role"`
}

func (r *dummyLoginRequest) validate() error {
	if r.Role != string(entity.RoleAdmin) && r.Role != string(entity.RoleUser) {
		return errors.New("role must be 'admin' or 'user'")
	}
	return nil
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (r *registerRequest) validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if r.Role != string(entity.RoleAdmin) && r.Role != string(entity.RoleUser) {
		return errors.New("role must be 'admin' or 'user'")
	}
	return nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *loginRequest) validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type userDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func userToDTO(u entity.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
}

type userResponse struct {
	User userDTO `json:"user"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

// ── Rooms ────────────────────────────────────────────────────────────────────

type createRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Capacity    *int    `json:"capacity"`
}

func (r *createRoomRequest) validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type roomDTO struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Capacity    *int       `json:"capacity,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
}

func roomToDTO(r entity.Room) roomDTO {
	t := r.CreatedAt
	return roomDTO{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Capacity:    r.Capacity,
		CreatedAt:   &t,
	}
}

type roomResponse struct {
	Room roomDTO `json:"room"`
}

type roomListResponse struct {
	Rooms []roomDTO `json:"rooms"`
}

// ── Schedules ────────────────────────────────────────────────────────────────

type createScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

func (r *createScheduleRequest) validate() error {
	if len(r.DaysOfWeek) == 0 {
		return errors.New("daysOfWeek is required and must not be empty")
	}
	for _, day := range r.DaysOfWeek {
		if day < 1 || day > 7 {
			return errors.New("each daysOfWeek value must be between 1 and 7")
		}
	}
	if r.StartTime == "" {
		return errors.New("startTime is required")
	}
	if r.EndTime == "" {
		return errors.New("endTime is required")
	}
	return nil
}

type scheduleDTO struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

func scheduleToDTO(s entity.Schedule) scheduleDTO {
	return scheduleDTO{
		ID:         s.ID,
		RoomID:     s.RoomID,
		DaysOfWeek: s.DaysOfWeek,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
	}
}

type scheduleResponse struct {
	Schedule scheduleDTO `json:"schedule"`
}

// ── Slots ────────────────────────────────────────────────────────────────────

type slotDTO struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"roomId"`
	StartTime time.Time `json:"start"`
	EndTime   time.Time `json:"end"`
}

func slotToDTO(s entity.Slot) slotDTO {
	return slotDTO{
		ID:        s.ID,
		RoomID:    s.RoomID,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
	}
}

type slotListResponse struct {
	Slots []slotDTO `json:"slots"`
}

// ── Bookings ─────────────────────────────────────────────────────────────────

type createBookingRequest struct {
	SlotID               uuid.UUID `json:"slotId"`
	CreateConferenceLink bool      `json:"createConferenceLink"`
}

func (r *createBookingRequest) validate() error {
	if r.SlotID == uuid.Nil {
		return errors.New("slotId is required")
	}
	return nil
}

type bookingDTO struct {
	ID             uuid.UUID `json:"id"`
	SlotID         uuid.UUID `json:"slotId"`
	UserID         uuid.UUID `json:"userId"`
	Status         string    `json:"status"`
	ConferenceLink *string   `json:"conferenceLink,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

func bookingToDTO(b entity.Booking) bookingDTO {
	return bookingDTO{
		ID:             b.ID,
		SlotID:         b.SlotID,
		UserID:         b.UserID,
		Status:         string(b.Status),
		ConferenceLink: b.ConferenceLink,
		CreatedAt:      b.CreatedAt,
	}
}

type bookingResponse struct {
	Booking bookingDTO `json:"booking"`
}

type bookingListResponse struct {
	Bookings   []bookingDTO   `json:"bookings"`
	Pagination *paginationDTO `json:"pagination,omitempty"`
}

type paginationDTO struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}
