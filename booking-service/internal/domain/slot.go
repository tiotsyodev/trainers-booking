package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID        uuid.UUID
	ClientID  uuid.UUID
	TrainerID uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

type SlotRepo interface {
	BookSlot(ctx context.Context, clientID, trainerID uuid.UUID, timeRange *TimeRange, ev SlotBooked) (*Slot, error)
	ListBooked(ctx context.Context, trainerID uuid.UUID, from, to time.Time) ([]TimeRange, error)
	Cancel(ctx context.Context, slotID uuid.UUID) error
}
