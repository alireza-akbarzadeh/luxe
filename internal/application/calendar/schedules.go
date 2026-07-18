package calendar

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// CreateSchedule creates a store's working schedule.
func (s *Service) CreateSchedule(ctx context.Context, req *dto.CreateStoreWorkingScheduleRequest) (*dto.StoreWorkingScheduleResponse, error) {
	if req.OpenTime != nil {
		if err := s.domain.ValidateTimeOfDay(*req.OpenTime); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.CloseTime != nil {
		if err := s.domain.ValidateTimeOfDay(*req.CloseTime); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	if existing, err := s.scheduleRepo.GetByStoreID(ctx, req.StoreID); err == nil && existing != nil {
		return nil, utils.ErrBadRequest("a working schedule already exists for this store; use update instead")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	sched, err := buildScheduleCreateModel(req)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid working_days payload")
	}
	if err := s.scheduleRepo.Create(ctx, sched); err != nil {
		return nil, err
	}
	return toScheduleResponse(sched), nil
}

// UpdateSchedule updates an existing working schedule by id.
func (s *Service) UpdateSchedule(ctx context.Context, id uint, req *dto.UpdateStoreWorkingScheduleRequest) (*dto.StoreWorkingScheduleResponse, error) {
	if req.OpenTime != nil {
		if err := s.domain.ValidateTimeOfDay(*req.OpenTime); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.CloseTime != nil {
		if err := s.domain.ValidateTimeOfDay(*req.CloseTime); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	sched, err := s.scheduleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("working schedule not found")
		}
		return nil, err
	}

	if err := applyScheduleUpdate(sched, req); err != nil {
		return nil, utils.ErrBadRequest("invalid working_days payload")
	}
	if err := s.scheduleRepo.Save(ctx, sched); err != nil {
		return nil, err
	}
	return toScheduleResponse(sched), nil
}

// ListSchedules returns working schedules, optionally filtered by store.
func (s *Service) ListSchedules(ctx context.Context, storeID *uint) ([]dto.StoreWorkingScheduleResponse, error) {
	schedules, err := s.scheduleRepo.List(ctx, storeID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.StoreWorkingScheduleResponse, 0, len(schedules))
	for i := range schedules {
		resp = append(resp, *toScheduleResponse(&schedules[i]))
	}
	return resp, nil
}

// GetScheduleByStoreID returns a store's working schedule, or nil if none is configured.
func (s *Service) GetScheduleByStoreID(ctx context.Context, storeID uint) (*dto.StoreWorkingScheduleResponse, error) {
	sched, err := s.scheduleRepo.GetByStoreID(ctx, storeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toScheduleResponse(sched), nil
}
