package entity

import "github.com/google/uuid"

type Room struct {
	ID          uuid.UUID
	Name        string
	Description string
	Capacity    int
}
