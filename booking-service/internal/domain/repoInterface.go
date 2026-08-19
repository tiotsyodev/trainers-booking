package domain

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, name, specialization string) (*Trainer, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Trainer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]Trainer, error)
}
