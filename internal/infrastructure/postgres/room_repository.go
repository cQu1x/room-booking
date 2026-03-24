package postgres

import (
	"context"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

// NewRoomRepository создаёт репозиторий переговорок на основе пула соединений.
func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

// CreateRoom сохраняет переговорку и возвращает созданную запись.
func (r *RoomRepository) CreateRoom(ctx context.Context, room *entity.Room) (*entity.Room, error) {
	const query = `INSERT INTO rooms (id, name, description, capacity, created_at)
	 VALUES ($1, $2, $3, $4, $5) RETURNING id, name, description, capacity, created_at`
	row := r.db.QueryRow(ctx, query, room.ID, room.Name, room.Description, room.Capacity, room.CreatedAt)
	createdRoom, err := scanRoom(row)
	if err != nil {
		return nil, err
	}
	return createdRoom, nil
}

// GetRoomByID возвращает переговорку по идентификатору.
func (r *RoomRepository) GetRoomByID(ctx context.Context, id uuid.UUID) (*entity.Room, error) {
	const query = `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	room, err := scanRoom(row)
	if err != nil {
		return nil, err
	}
	return room, nil
}

// ListRooms возвращает список всех переговорок.
func (r *RoomRepository) ListRooms(ctx context.Context) ([]entity.Room, error) {
	const query = `SELECT id, name, description, capacity, created_at FROM rooms`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []entity.Room

	for rows.Next() {
		var room entity.Room
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rooms, nil
}

func scanRoom(row pgx.Row) (*entity.Room, error) {
	var room entity.Room
	err := row.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &room, nil
}
