package domain

import "github.com/google/uuid"

type Trainer struct {
	ID             uuid.UUID
	Specialization string
	Name           string
}
