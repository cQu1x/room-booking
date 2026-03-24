package postgres

import (
	"context"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepository struct {
	db *pgxpool.Pool
}

func NewSlotRepository(db *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{db: db}
}

func (r *SlotRepository) CreateSlots(ctx context.Context, slots []entity.Slot) error {
	if len(slots) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, slot := range slots {
		batch.Queue(
			`INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
			slot.ID, slot.RoomID, slot.StartTime, slot.EndTime,
		)
	}
	br := r.db.SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()
	for range slots {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *SlotRepository) GetMaxSlotDate(ctx context.Context, roomID uuid.UUID) (*time.Time, error) {
	const query = `SELECT MAX(start_time) FROM slots WHERE room_id = $1`
	var maxDate *time.Time
	if err := r.db.QueryRow(ctx, query, roomID).Scan(&maxDate); err != nil {
		return nil, err
	}
	return maxDate, nil
}

func (r *SlotRepository) GetSlotByID(ctx context.Context, id uuid.UUID) (*entity.Slot, error) {
	const query = `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	var slot entity.Slot
	if err := r.db.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime); err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *SlotRepository) ListAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]entity.Slot, error) {
	const query = `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		WHERE s.room_id = $1
		  AND s.start_time::date = $2::date
		  AND NOT EXISTS (
		    SELECT 1 FROM bookings b
		    WHERE b.slot_id = s.id AND b.status = 'active'
		  )
		ORDER BY s.start_time`
	rows, err := r.db.Query(ctx, query, roomID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []entity.Slot
	for rows.Next() {
		var slot entity.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return slots, nil
}
