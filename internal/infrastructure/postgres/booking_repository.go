package postgres

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

// NewBookingRepository создаёт репозиторий бронирований на основе пула соединений.
func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

// CreateBooking сохраняет новое бронирование в базе данных.
func (r *BookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	const query = `INSERT INTO bookings (id, user_id, slot_id, status, conference_link, created_at)
	 VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query,
		booking.ID, booking.UserID, booking.SlotID, booking.Status, booking.ConferenceLink, booking.CreatedAt,
	)
	return err
}

// GetBookingByID возвращает бронирование по идентификатору.
func (r *BookingRepository) GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Booking, error) {
	const query = `SELECT id, user_id, slot_id, status, conference_link, created_at FROM bookings WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	return scanBooking(row)
}

// CancelBooking переводит статус бронирования в «отменено».
func (r *BookingRepository) CancelBooking(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, entity.BookingStatusCancelled, id)
	return err
}

// ListAllBookings возвращает постраничный список всех бронирований и общее количество записей.
func (r *BookingRepository) ListAllBookings(ctx context.Context, page, pageSize int) ([]entity.Booking, int, error) {
	const countQuery = `SELECT COUNT(*) FROM bookings`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `SELECT id, user_id, slot_id, status, conference_link, created_at
	 FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []entity.Booking
	for rows.Next() {
		var b entity.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}

// ListByUserID возвращает будущие бронирования пользователя, отсортированные по времени начала слота.
func (r *BookingRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Booking, error) {
	const query = `
		SELECT b.id, b.user_id, b.slot_id, b.status, b.conference_link, b.created_at
		FROM bookings b
		JOIN slots s ON s.id = b.slot_id
		WHERE b.user_id = $1
		  AND s.start_time > NOW()
		ORDER BY s.start_time`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.Booking
	for rows.Next() {
		var b entity.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

// IsSlotBooked возвращает true, если на слот есть активное бронирование.
func (r *BookingRepository) IsSlotBooked(ctx context.Context, slotID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id = $1 AND status = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, slotID, entity.BookingStatusActive).Scan(&exists)
	return exists, err
}

func scanBooking(row pgx.Row) (*entity.Booking, error) {
	var b entity.Booking
	if err := row.Scan(&b.ID, &b.UserID, &b.SlotID, &b.Status, &b.ConferenceLink, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}
