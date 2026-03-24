package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	jwtpkg "github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	dummyAdminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type AuthUseCase struct {
	userRepo     ports.UserRepository
	tokenManager *jwtpkg.TokenManager
}

// NewAuthUseCase создаёт сервис аутентификации.
func NewAuthUseCase(userRepo ports.UserRepository, tokenManager *jwtpkg.TokenManager) *AuthUseCase {
	return &AuthUseCase{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

// DummyLogin выдаёт JWT с фиксированным user_id для указанной роли (admin или user).
func (a *AuthUseCase) DummyLogin(ctx context.Context, role entity.Role) (string, error) {
	var userID uuid.UUID
	switch role {
	case entity.RoleAdmin:
		userID = dummyAdminID
	case entity.RoleUser:
		userID = dummyUserID
	default:
		return "", fmt.Errorf("unknown role: %s", role)
	}
	return a.tokenManager.GenerateToken(userID, role)
}

// Register регистрирует нового пользователя с хэшированным паролем.
func (a *AuthUseCase) Register(ctx context.Context, email, password string, role entity.Role) (*entity.User, error) {
	existing, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return nil, entity.ErrEmailAlreadyTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now().UTC(),
	}

	created, err := a.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// Login проверяет учётные данные и возвращает подписанный JWT при успехе.
func (a *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", entity.ErrInvalidCredentials
	}

	return a.tokenManager.GenerateToken(user.ID, user.Role)
}
