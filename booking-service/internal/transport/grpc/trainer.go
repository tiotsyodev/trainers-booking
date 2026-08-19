package grpc

import (
	trainerv1 "booker/gen/trainer/v1"
	"booker/booking-service/internal/domain"
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TrainerHandler struct {
	trainerv1.UnimplementedTrainerServiceServer
	repo   domain.Repository
	whrepo domain.WorkingHoursRepo
}

func NewTrainerHandler(repo domain.Repository, whrepo domain.WorkingHoursRepo) *TrainerHandler {
	return &TrainerHandler{
		repo:   repo,
		whrepo: whrepo,
	}
}

func (h *TrainerHandler) CreateTrainer(ctx context.Context, req *trainerv1.CreateTrainerRequest) (*trainerv1.CreateTrainerResponse, error) {
	trainer, err := h.repo.Create(ctx, req.Name, req.Specialization)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "trainer with this name already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to create trainer")
		}
	}

	return &trainerv1.CreateTrainerResponse{
		Trainer: &trainerv1.Trainer{
			Id:             trainer.ID.String(),
			Name:           trainer.Name,
			Specialization: trainer.Specialization,
		},
	}, nil
}

func (h *TrainerHandler) GetTrainer(ctx context.Context, req *trainerv1.GetTrainerRequest) (*trainerv1.GetTrainerResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	trainer, err := h.repo.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "trainer not found")
		default:
			return nil, status.Error(codes.Internal, "failed to get trainer")
		}
	}

	return &trainerv1.GetTrainerResponse{
		Trainer: &trainerv1.Trainer{
			Id:             trainer.ID.String(),
			Name:           trainer.Name,
			Specialization: trainer.Specialization,
		},
	}, nil
}

func (h *TrainerHandler) GetAllTrainers(ctx context.Context, req *trainerv1.GetAllTrainersRequest) (*trainerv1.GetAllTrainersResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	offset := 0
	if req.PageToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.PageToken)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid page token")
		}
		offset, err = strconv.Atoi(string(decoded))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid page token")
		}
	}

	trainers, err := h.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list trainers")
	}

	resp := &trainerv1.GetAllTrainersResponse{}
	for _, t := range trainers {
		resp.Trainers = append(resp.Trainers, &trainerv1.Trainer{
			Id: t.ID.String(), Name: t.Name, Specialization: t.Specialization,
		})
	}

	if len(trainers) == pageSize {
		nextOffset := offset + pageSize
		resp.NextPageToken = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(nextOffset)))
	}

	return resp, nil
}

func (h *TrainerHandler) DeleteTrainer(ctx context.Context, req *trainerv1.DeleteTrainerRequest) (*trainerv1.DeleteTrainerResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "trainer not found")
		default:
			return nil, status.Error(codes.Internal, "failed to delete trainer")
		}
	}
	return &trainerv1.DeleteTrainerResponse{}, nil
}

func (h *TrainerHandler) SetWorkingHours(ctx context.Context, req *trainerv1.SetWorkingHoursRequest) (*trainerv1.SetWorkingHoursResponse, error) {
	trainerID, err := uuid.Parse(req.TrainerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trainer id")
	}

	hours := make([]domain.WorkingHours, 0, len(req.Days))
	seen := make(map[int32]bool)

	for _, d := range req.Days {
		_, ok := seen[d.DayOfWeek]
		if ok {
			return nil, status.Error(codes.InvalidArgument, "double day of week write")
		}
		seen[d.DayOfWeek] = true

		wh := domain.WorkingHours{
			TrainerID:              trainerID,
			DayOfWeek:              time.Weekday(d.DayOfWeek),
			StartTime:              d.StartTime,
			EndTime:                d.EndTime,
			SessionDurationMinutes: int(d.SessionDurationMinutes),
		}

		if err := wh.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		hours = append(hours, wh)
	}

	if err := h.whrepo.SetWorkingHours(ctx, trainerID, hours); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "trainer not found")
		default:
			return nil, status.Error(codes.Internal, "failed to set working hours")
		}
	}

	return &trainerv1.SetWorkingHoursResponse{
		Days: req.Days,
	}, nil

}

func (h *TrainerHandler) GetWorkingHours(ctx context.Context, req *trainerv1.GetWorkingHoursRequest) (*trainerv1.GetWorkingHoursResponse, error) {
	trainerID, err := uuid.Parse(req.TrainerId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	days, err := h.whrepo.GetWorkingHours(ctx, trainerID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get working hours")
	}

	schedule := make([]*trainerv1.DaySchedule, 0, len(days))

	for _, v := range days {
		schedule = append(schedule, &trainerv1.DaySchedule{
			DayOfWeek:              int32(v.DayOfWeek),
			StartTime:              v.StartTime,
			EndTime:                v.EndTime,
			SessionDurationMinutes: int32(v.SessionDurationMinutes),
		})

	}

	return &trainerv1.GetWorkingHoursResponse{
		Days: schedule,
	}, nil
}
