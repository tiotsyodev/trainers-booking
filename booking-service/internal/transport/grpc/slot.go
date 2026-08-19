package grpc

import (
	slotv1 "booker/gen/slot/v1"
	"booker/booking-service/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SlotSvc interface {
	ListAvailable(ctx context.Context, trainerID uuid.UUID, date time.Time) ([]domain.TimeRange, error)
	Book(ctx context.Context, clientID, trainerID uuid.UUID, startTime time.Time) (*domain.Slot, error)
	GetCalendar(ctx context.Context, trainerID uuid.UUID, from, to time.Time) ([]domain.DayAvailability, error)
	CancelSlot(ctx context.Context, slotID uuid.UUID) error
}

type SlotHandler struct {
	slotv1.UnimplementedSlotServiceServer
	svc SlotSvc
}

func NewSlotHandler(svc SlotSvc) *SlotHandler {
	return &SlotHandler{
		svc: svc,
	}
}

func (h *SlotHandler) BookSlot(ctx context.Context, req *slotv1.BookSlotReq) (*slotv1.BookSlotRes, error) {

	clientID, err := uuid.Parse(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client id")
	}

	trainerID, err := uuid.Parse(req.TrainerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trainer id")
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "start_time should be RFC3339 format")
	}

	slot, err := h.svc.Book(ctx, clientID, trainerID, start)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "slot is already busy")
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "slot is not found")
		case errors.Is(err, domain.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Errorf(codes.Internal, "failed to book slot")
	}

	return &slotv1.BookSlotRes{
		Slot: &slotv1.Slot{
			Id:        slot.ID.String(),
			ClientId:  slot.ClientID.String(),
			TrainerId: slot.TrainerID.String(),
			StartTime: slot.StartTime.Format(time.RFC3339),
			EndTime:   slot.EndTime.Format(time.RFC3339),
			Status:    slot.Status,
		},
	}, nil
}

func (h *SlotHandler) ListAvailableSlots(ctx context.Context, req *slotv1.ListAvailableSlotsReq) (*slotv1.ListAvailableSlotsRes, error) {

	trainerID, err := uuid.Parse(req.TrainerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trainer id")
	}

	date, err := time.Parse(time.DateOnly, req.Date)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	ranges, err := h.svc.ListAvailable(ctx, trainerID, date)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list available slots")
	}

	ansRanges := make([]*slotv1.TimeRange, 0, len(ranges))

	for _, v := range ranges {
		ansRanges = append(ansRanges, &slotv1.TimeRange{
			Start: v.Start.Format(time.RFC3339),
			End:   v.End.Format(time.RFC3339),
		})
	}

	return &slotv1.ListAvailableSlotsRes{
		TimeRange: ansRanges,
	}, nil
}

func (h *SlotHandler) GetCalendar(ctx context.Context, req *slotv1.GetAvailabilityCalendarReq) (*slotv1.GetAvailabilityCalendarRes, error) {

	trainerID, err := uuid.Parse(req.TrainerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trainer id")
	}

	dateFrom, err := time.Parse(time.DateOnly, req.DateFrom)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date_from")
	}

	dateTo, err := time.Parse(time.DateOnly, req.DateTo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date_to")
	}

	days, err := h.svc.GetCalendar(ctx, trainerID, dateFrom, dateTo)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to get calendar")

	}

	res := make([]*slotv1.DayAvailability, 0, len(days))

	for _, v := range days {
		res = append(res, &slotv1.DayAvailability{
			Date:      v.Date.Format(time.DateOnly),
			Available: v.Available,
		})
	}

	return &slotv1.GetAvailabilityCalendarRes{
		Days: res,
	}, nil
}

func (h *SlotHandler) CancelSlot(ctx context.Context, req *slotv1.CancelSlotReq) (*slotv1.CancelSlotRes, error) {
	slotID, err := uuid.Parse(req.SlotId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid slot id")
	}

	if err := h.svc.CancelSlot(ctx, slotID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "busy slot not found")
		}

		return nil, status.Error(codes.Internal, "failed to cancel slot")
	}

	return &slotv1.CancelSlotRes{}, nil
}

var _ slotv1.SlotServiceServer = (*SlotHandler)(nil)
