package postgres

import (
	"booker/trainer-service/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) GetWorkingHours(ctx context.Context, trainerID uuid.UUID) ([]domain.WorkingHours, error) {
	query := "select trainer_id, day_of_week, to_char(start_time, 'HH24:MI'), to_char(end_time, 'HH24:MI'), session_duration_minutes from working_hours where trainer_id=$1"

	rows, err := r.Query(ctx, query, trainerID)
	if err != nil {
		return nil, fmt.Errorf("unable to query working hours: %w", err)
	}
	defer rows.Close()

	var hours []domain.WorkingHours
	for rows.Next() {
		var hour domain.WorkingHours

		if err := rows.Scan(&hour.TrainerID, &hour.DayOfWeek, &hour.StartTime, &hour.EndTime, &hour.SessionDurationMinutes); err != nil {
			return nil, fmt.Errorf("unable to scan working hours: %w", err)
		}

		hours = append(hours, hour)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("hours iteration error: %w", err)
	}

	return hours, nil
}

func (r *Repo) SetWorkingHours(ctx context.Context, trainerID uuid.UUID, workingHorus []domain.WorkingHours) error {
	tx, err := r.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begint tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "delete from working_hours where trainer_id=$1", trainerID); err != nil {
		return fmt.Errorf("deleting existing working hours: %w", err)
	}

	for _, h := range workingHorus {
		_, err := tx.Exec(ctx, "insert into working_hours (trainer_id, day_of_week, start_time, end_time, session_duration_minutes) values ($1, $2, $3::time, $4::time, $5)", trainerID, int(h.DayOfWeek), h.StartTime, h.EndTime, h.SessionDurationMinutes)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return domain.ErrNotFound
			}
			return fmt.Errorf("create hour: %w", err)
		}
	}

	return tx.Commit(ctx)
}
