package room_test

import (
	"context"
	"errors"
	"testing"

	"github.com/avito-internships/test-backend-1-cQu1x/internal/domain/entity"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/testutil"
	"github.com/avito-internships/test-backend-1-cQu1x/internal/usecase/room"
	"github.com/google/uuid"
)

func TestCreate_Success(t *testing.T) {
	name := "Conference A"
	desc := "Main floor"
	cap := 10

	roomRepo := &testutil.MockRoomRepo{
		CreateRoomFn: func(_ context.Context, r *entity.Room) (*entity.Room, error) {
			return r, nil
		},
	}

	svc := room.NewService(roomRepo)
	created, err := svc.Create(context.Background(), name, &desc, &cap)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Name != name {
		t.Errorf("expected name %q, got %q", name, created.Name)
	}
	if created.Description == nil || *created.Description != desc {
		t.Errorf("expected description %q", desc)
	}
	if created.Capacity == nil || *created.Capacity != cap {
		t.Errorf("expected capacity %d", cap)
	}
	if created.ID == uuid.Nil {
		t.Error("expected a non-nil UUID to be assigned")
	}
}

func TestCreate_WithoutOptionalFields(t *testing.T) {
	roomRepo := &testutil.MockRoomRepo{
		CreateRoomFn: func(_ context.Context, r *entity.Room) (*entity.Room, error) {
			return r, nil
		},
	}

	svc := room.NewService(roomRepo)
	created, err := svc.Create(context.Background(), "Room B", nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Description != nil {
		t.Error("expected nil description")
	}
	if created.Capacity != nil {
		t.Error("expected nil capacity")
	}
}

func TestCreate_RepositoryError(t *testing.T) {
	repoErr := errors.New("db error")
	roomRepo := &testutil.MockRoomRepo{
		CreateRoomFn: func(_ context.Context, _ *entity.Room) (*entity.Room, error) {
			return nil, repoErr
		},
	}

	svc := room.NewService(roomRepo)
	_, err := svc.Create(context.Background(), "Room C", nil, nil)

	if !errors.Is(err, repoErr) {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestList_Success(t *testing.T) {
	expected := []entity.Room{
		{ID: uuid.New(), Name: "Room 1"},
		{ID: uuid.New(), Name: "Room 2"},
	}
	roomRepo := &testutil.MockRoomRepo{
		ListRoomsFn: func(_ context.Context) ([]entity.Room, error) {
			return expected, nil
		},
	}

	svc := room.NewService(roomRepo)
	rooms, err := svc.List(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != len(expected) {
		t.Errorf("expected %d rooms, got %d", len(expected), len(rooms))
	}
}

func TestGetByID_Success(t *testing.T) {
	id := uuid.New()
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, rid uuid.UUID) (*entity.Room, error) {
			return &entity.Room{ID: rid, Name: "Meeting Room"}, nil
		},
	}

	svc := room.NewService(roomRepo)
	r, err := svc.GetByID(context.Background(), id)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != id {
		t.Errorf("expected ID %v, got %v", id, r.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	roomRepo := &testutil.MockRoomRepo{
		GetRoomByIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Room, error) {
			return nil, entity.ErrRoomNotFound
		},
	}

	svc := room.NewService(roomRepo)
	_, err := svc.GetByID(context.Background(), uuid.New())

	if !errors.Is(err, entity.ErrRoomNotFound) {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}
