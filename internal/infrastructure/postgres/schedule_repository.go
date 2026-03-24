package postgres

import (
	"context"
	"errors"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	db *pgxpool.Pool
}

func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) CreateSchedule(ctx context.Context, schedule *entity.Schedule) error {
	const query = `INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
	 VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime)
	return err
}

func (r *ScheduleRepository) GetScheduleByRoomID(ctx context.Context, roomID uuid.UUID) (*entity.Schedule, error) {
	const query = `SELECT id, room_id, days_of_week, start_time, end_time FROM schedules WHERE room_id = $1`
	row := r.db.QueryRow(ctx, query, roomID)
	var s entity.Schedule
	err := row.Scan(&s.ID, &s.RoomID, &s.DaysOfWeek, &s.StartTime, &s.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}
