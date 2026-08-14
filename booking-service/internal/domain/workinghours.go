package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkingHours struct {
	TrainerID              uuid.UUID
	DayOfWeek              time.Weekday
	StartTime              string
	EndTime                string
	SessionDurationMinutes int
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type WorkingHoursRepo interface {
	SetWorkingHours(ctx context.Context, trainerID uuid.UUID, hours []WorkingHours) error
	GetWorkingHours(ctx context.Context, trainerID uuid.UUID) ([]WorkingHours, error)
}

func (w *WorkingHours) Validate() error {
	if w.DayOfWeek < time.Sunday || w.DayOfWeek > time.Saturday {
		return fmt.Errorf("%w: day_of_week must be 0..6", ErrInvalidArgument)
	}

	start, err := time.Parse("15:04", w.StartTime)
	if err != nil {
		return fmt.Errorf("%w: start_time must be HH:MM", ErrInvalidArgument)
	}
	end, err := time.Parse("15:04", w.EndTime)
	if err != nil {
		return fmt.Errorf("%w: end_time must be HH:MM", ErrInvalidArgument)
	}
	if !start.Before(end) {
		return fmt.Errorf("%w: start_time must be before end_time", ErrInvalidArgument)
	}

	if w.SessionDurationMinutes < 15 || w.SessionDurationMinutes > 240 {
		return fmt.Errorf("%w: session_duration_minutes must be 15..240", ErrInvalidArgument)
	}

	return nil
}

func (w *WorkingHours) GenerateCandidates(date time.Time) ([]TimeRange, error) {
	start, err := time.Parse("15:04", w.StartTime)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse start time")
	}
	end, err := time.Parse("15:04", w.EndTime)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse end time")
	}

	y, m, d := date.Date()
	cur := time.Date(y, m, d, start.Hour(), start.Minute(), 0, 0, time.UTC)
	limit := time.Date(y, m, d, end.Hour(), end.Minute(), 0, 0, time.UTC)

	dur := time.Duration(w.SessionDurationMinutes) * time.Minute

	var candidates []TimeRange
	for !cur.Add(dur).After(limit) {
		candidates = append(candidates, TimeRange{Start: cur, End: cur.Add(dur)})
		cur = cur.Add(dur)
	}

	return candidates, nil
}
