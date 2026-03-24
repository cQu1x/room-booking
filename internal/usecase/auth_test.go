package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	jwtpkg "github.com/avito-internships/test-backend-1-cQu1x/internal/infrastructure/jwt"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	dummyAdminID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	dummyUserID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func newTokenManager() *jwtpkg.TokenManager {
	return jwtpkg.NewTokenManager("test-secret")
}

func TestDummyLogin_AdminReturnsFixedID(t *testing.T) {
	tm := newTokenManager()
	svc := usecase.NewAuthUseCase(&testutil.MockUserRepo{}, tm)

	token, err := svc.DummyLogin(context.Background(), entity.RoleAdmin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.UserID != dummyAdminID {
		t.Errorf("expected admin ID %v, got %v", dummyAdminID, claims.UserID)
	}
	if claims.Role != entity.RoleAdmin {
		t.Errorf("expected role admin, got %v", claims.Role)
	}
}

func TestDummyLogin_UserReturnsFixedID(t *testing.T) {
	tm := newTokenManager()
	svc := usecase.NewAuthUseCase(&testutil.MockUserRepo{}, tm)

	token, err := svc.DummyLogin(context.Background(), entity.RoleUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.UserID != dummyUserID {
		t.Errorf("expected user ID %v, got %v", dummyUserID, claims.UserID)
	}
	if claims.Role != entity.RoleUser {
		t.Errorf("expected role user, got %v", claims.Role)
	}
}

func TestDummyLogin_InvalidRoleReturnsError(t *testing.T) {
	svc := usecase.NewAuthUseCase(&testutil.MockUserRepo{}, newTokenManager())
	_, err := svc.DummyLogin(context.Background(), "superuser")
	if err == nil {
		t.Error("expected error for unknown role")
	}
}

func TestRegister_Success(t *testing.T) {
	userRepo := &testutil.MockUserRepo{
		GetUserByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, nil // no existing user
		},
		CreateUserFn: func(_ context.Context, u *entity.User) (entity.User, error) {
			return *u, nil
		},
	}

	svc := usecase.NewAuthUseCase(userRepo, newTokenManager())
	user, err := svc.Register(context.Background(), "test@example.com", "password123", entity.RoleUser)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("unexpected email: %v", user.Email)
	}
	if user.Role != entity.RoleUser {
		t.Errorf("unexpected role: %v", user.Role)
	}
	// Password must be stored hashed, not in plaintext.
	if user.PasswordHash == "password123" {
		t.Error("password must be hashed, not stored as plaintext")
	}
}

func TestRegister_EmailAlreadyTaken(t *testing.T) {
	existing := &entity.User{ID: uuid.New(), Email: "taken@example.com"}
	userRepo := &testutil.MockUserRepo{
		GetUserByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return existing, nil
		},
	}

	svc := usecase.NewAuthUseCase(userRepo, newTokenManager())
	_, err := svc.Register(context.Background(), "taken@example.com", "pass", entity.RoleUser)

	if !errors.Is(err, entity.ErrEmailAlreadyTaken) {
		t.Errorf("expected ErrEmailAlreadyTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	id := uuid.New()
	userRepo := &testutil.MockUserRepo{
		GetUserByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{
				ID:           id,
				Email:        "user@example.com",
				PasswordHash: string(hash),
				Role:         entity.RoleUser,
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	tm := newTokenManager()
	svc := usecase.NewAuthUseCase(userRepo, tm)
	token, err := svc.Login(context.Background(), "user@example.com", "secret")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("token invalid: %v", err)
	}
	if claims.UserID != id {
		t.Errorf("expected userID %v, got %v", id, claims.UserID)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &testutil.MockUserRepo{
		GetUserByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return nil, pgx.ErrNoRows
		},
	}

	svc := usecase.NewAuthUseCase(userRepo, newTokenManager())
	_, err := svc.Login(context.Background(), "nobody@example.com", "pass")

	if !errors.Is(err, entity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	userRepo := &testutil.MockUserRepo{
		GetUserByEmailFn: func(_ context.Context, _ string) (*entity.User, error) {
			return &entity.User{
				ID:           uuid.New(),
				Email:        "user@example.com",
				PasswordHash: string(hash),
				Role:         entity.RoleUser,
			}, nil
		},
	}

	svc := usecase.NewAuthUseCase(userRepo, newTokenManager())
	_, err := svc.Login(context.Background(), "user@example.com", "wrong")

	if !errors.Is(err, entity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
