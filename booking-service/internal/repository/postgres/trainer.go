package postgres

import (
	"booker/booking-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trainer, error) {
	query := `SELECT id , specialization, name FROM trainer WHERE id = $1`

	var trainer domain.Trainer
	if err := r.QueryRow(ctx, query, id).Scan(&trainer.ID, &trainer.Specialization, &trainer.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("unable to scan trainer: %w", err)
	}

	return &trainer, nil
}

func (r *Repo) List(ctx context.Context, limit, offset int) ([]domain.Trainer, error) {
	query := `select id, name, specialization from trainer order by created_at, id limit $1 offset $2`

	rows, err := r.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("unable to query trainers: %w", err)
	}
	defer rows.Close()

	var trainers []domain.Trainer
	for rows.Next() {
		var trainer domain.Trainer
		if err := rows.Scan(&trainer.ID, &trainer.Name, &trainer.Specialization); err != nil {
			return nil, fmt.Errorf("unable to scan trainer: %w", err)
		}
		trainers = append(trainers, trainer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return trainers, nil
}

func (r *Repo) Create(ctx context.Context, name, specialization string) (*domain.Trainer, error) {
	query := `INSERT INTO trainer (name, specialization) VALUES ($1, $2) RETURNING id, specialization, name`

	var trainer domain.Trainer
	if err := r.QueryRow(ctx, query, name, specialization).Scan(&trainer.ID, &trainer.Specialization, &trainer.Name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("unable to scan trainer: %w", err)
	}

	return &trainer, nil
}

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM trainer WHERE id = $1 RETURNING id`

	var deletedID uuid.UUID
	if err := r.QueryRow(ctx, query, id).Scan(&deletedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("unable to scan trainer: %w", err)
	}

	return nil
}
