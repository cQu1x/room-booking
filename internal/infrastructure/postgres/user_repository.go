package postgres

import (
	"context"
	"errors"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (entity.User, error) {
	const query = `INSERT INTO users (id, email, password_hash, role, created_at)
	 VALUES ($1, $2, $3, $4, $5) RETURNING id, email, password_hash, role, created_at`
	row := r.db.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt)
	return scanUser(row)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	const query = `SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func scanUser(row pgx.Row) (entity.User, error) {
	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return entity.User{}, err
	}
	return user, nil
}
