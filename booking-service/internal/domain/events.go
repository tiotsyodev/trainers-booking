package domain

import (
	"time"

	"github.com/google/uuid"
)

type SlotBooked struct {
	EventId    uuid.UUID
	ClientID   uuid.UUID
	TrainerID  uuid.UUID
	SlotID     uuid.UUID
	OccurredAt time.Time
	Start      time.Time
	End        time.Time
}
