package service

import (
	"booker/booking-service/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SlotSvc struct {
	whRepo   domain.WorkingHoursRepo
	slotRepo domain.SlotRepo
}

func NewSlotService(whRepo domain.WorkingHoursRepo, slotRepo domain.SlotRepo) *SlotSvc {
	return &SlotSvc{
		whRepo:   whRepo,
		slotRepo: slotRepo,
	}
}

func (avs *SlotSvc) ListAvailable(ctx context.Context, trainerID uuid.UUID, date time.Time) ([]domain.TimeRange, error) {

	wh, err := avs.whRepo.GetWorkingHours(ctx, trainerID)
	if err != nil {
		return nil, fmt.Errorf("get working hours: %w", err)
	}

	var rule *domain.WorkingHours

	for _, v := range wh {
		if v.DayOfWeek == date.Weekday() {
			rule = &v
			break
		}
	}

	if rule == nil {
		return nil, nil
	}

	candidates, err := rule.GenerateCandidates(date)
	if err != nil {
		return nil, fmt.Errorf("gen candidates: %w", err)
	}

	y, m, d := date.Date()
	dayStart := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	booked, err := avs.slotRepo.ListBooked(ctx, trainerID, dayStart, dayStart.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("get booked: %w", err)
	}

	bookedmp := make(map[int64]bool, len(booked))
	for _, v := range booked {
		bookedmp[v.Start.UTC().Unix()] = true
	}

	var available []domain.TimeRange
	for _, v := range candidates {
		if v.Start.Before(time.Now().Add(time.Hour)) {
			continue
		}
		if bookedmp[v.Start.UTC().Unix()] {
			continue
		}

		available = append(available, v)
	}
	return available, nil
}

func (avs *SlotSvc) GetCalendar(ctx context.Context, trainerID uuid.UUID, from, to time.Time) ([]domain.DayAvailability, error) {
	wh, err := avs.whRepo.GetWorkingHours(ctx, trainerID)
	if err != nil {
		return nil, fmt.Errorf("get working hours: %w", err)
	}

	from, to = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location()), time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location())

	if to.Sub(from) > 90*24*time.Hour {
		return nil, fmt.Errorf("%w: date range must not exceed 90 days", domain.ErrInvalidArgument)
	}

	if from.After(to) {
		return nil, fmt.Errorf("%w: date_from must be before date_to", domain.ErrInvalidArgument)
	}

	booked, err := avs.slotRepo.ListBooked(ctx, trainerID, from, to.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("get booked: %w", err)
	}

	bookedmp := make(map[int64]bool, len(booked))
	for _, v := range booked {
		bookedmp[v.Start.UTC().Unix()] = true
	}

	cutoff := time.Now().Add(time.Hour)

	var days []domain.DayAvailability
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		var rule *domain.WorkingHours

		for _, v := range wh {
			if v.DayOfWeek == d.Weekday() {
				rule = &v
				break
			}
		}

		if rule == nil {
			days = append(days, domain.DayAvailability{Date: d, Available: false})
			continue
		}

		candidates, err := rule.GenerateCandidates(d)
		if err != nil {
			return nil, fmt.Errorf("gen candidates: %w", err)
		}

		var available []domain.TimeRange
		for _, v := range candidates {
			if v.Start.Before(cutoff) {
				continue
			}
			if bookedmp[v.Start.UTC().Unix()] {
				continue
			}

			available = append(available, v)
		}

		days = append(days, domain.DayAvailability{Date: d, Available: len(available) > 0})

	}

	return days, nil
}

func (avs *SlotSvc) Book(ctx context.Context, clientID, trainerID uuid.UUID, startTime time.Time) (*domain.Slot, error) {

	wh, err := avs.whRepo.GetWorkingHours(ctx, trainerID)
	if err != nil {
		return nil, fmt.Errorf("get working hours: %w", err)
	}

	var rule *domain.WorkingHours
	for _, v := range wh {
		if v.DayOfWeek == startTime.Weekday() {
			rule = &v
		}
	}

	if startTime.Before(time.Now().Add(time.Hour)) {
		return nil, fmt.Errorf("%w: booking must be at least %v in advance", domain.ErrInvalidArgument, time.Hour)
	}

	if rule == nil {
		return nil, fmt.Errorf("%w: trainer does not work this day", domain.ErrInvalidArgument)
	}

	candidates, err := rule.GenerateCandidates(startTime)
	if err != nil {
		return nil, fmt.Errorf("generate candidates: %w", err)
	}

	var match *domain.TimeRange
	for _, v := range candidates {
		if v.Start.UTC().Unix() == startTime.UTC().Unix() {
			match = &v
		}
	}
	if match == nil {
		return nil, fmt.Errorf("%w: no such slot", domain.ErrInvalidArgument)
	}

	return avs.slotRepo.BookSlot(ctx, clientID, trainerID, match)
}

func (avs *SlotSvc) CancelSlot(ctx context.Context, slotID uuid.UUID) error {
	return avs.slotRepo.Cancel(ctx, slotID)
}
