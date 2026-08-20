package postgres

import (
	"booker/booking-service/internal/domain"
	eventsv1 "booker/gen/events/v1"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *Repo) BookSlot(ctx context.Context, clientID, trainerID uuid.UUID, timeRange *domain.TimeRange, ev domain.SlotBooked) (*domain.Slot, error) {

	pbEvent := eventsv1.Event{
		EventId:   ev.EventId.String(),
		OccuredAt: timestamppb.New(ev.OccurredAt),
		Payload: &eventsv1.Event_SlotBooked{
			SlotBooked: &eventsv1.SlotBooked{
				SlotId:    ev.SlotID.String(),
				TrainerId: ev.TrainerID.String(),
				ClientId:  ev.ClientID.String(),
				StartTime: timestamppb.New(ev.Start),
				EndTime:   timestamppb.New(ev.End),
			},
		},
	}

	data, err := proto.Marshal(&pbEvent)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	// Init tx
	tx, err := r.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Booking Slot
	bookQuery := "INSERT INTO slot (id, start_time, end_time, trainer_id, client_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, status"
	var slot domain.Slot
	err = tx.QueryRow(ctx, bookQuery, ev.SlotID, timeRange.Start, timeRange.End, trainerID, clientID).Scan(&slot.ID, &slot.Status)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case "23505":
				return nil, domain.ErrAlreadyExists
			case "23503":
				return nil, domain.ErrNotFound
			}
		}
		return nil, fmt.Errorf("unable to book slot: %w", err)
	}

	// Create Event
	eventQuery := "INSERT INTO outbox_events (id, partition_key, payload) VALUES ($1, $2, $3)"
	_, err = tx.Exec(ctx, eventQuery, ev.EventId, ev.SlotID.String(), data)
	if err != nil {
		return nil, fmt.Errorf("save outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	slot.ClientID = clientID
	slot.TrainerID = trainerID
	slot.StartTime = timeRange.Start
	slot.EndTime = timeRange.End

	return &slot, nil

}

func (r *Repo) Cancel(ctx context.Context, slotID uuid.UUID) error {
	query := "UPDATE slot SET status='cancelled' where id=$1 and status='active'"

	tag, err := r.Exec(ctx, query, slotID)
	if err != nil {
		return fmt.Errorf("unable to cancel slot: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *Repo) ListBooked(ctx context.Context, trainerID uuid.UUID, from, to time.Time) ([]domain.TimeRange, error) {
	query := "SELECT start_time, end_time FROM slot WHERE start_time >= $1 and start_time < $2 AND status='active' AND trainer_id=$3 ORDER BY start_time"

	var timeRanges []domain.TimeRange

	rows, err := r.Query(ctx, query, from, to, trainerID)
	if err != nil {
		return nil, fmt.Errorf("unable to querry slots: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var timeRange domain.TimeRange

		err := rows.Scan(&timeRange.Start, &timeRange.End)
		if err != nil {
			return nil, fmt.Errorf("unable to scan slot: %w", err)
		}

		timeRanges = append(timeRanges, timeRange)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("slot iteration error: %w", err)
	}

	return timeRanges, nil
}

var _ domain.SlotRepo = (*Repo)(nil)
