package calendar

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// CreateHoliday creates a holiday/closure window after domain validation.
func (s *Service) CreateHoliday(ctx context.Context, req *dto.CreateStoreHolidayRequest, createdBy *uint) (*dto.StoreHolidayResponse, error) {
	if err := s.domain.ValidateName(req.Name); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := s.domain.ValidateHolidayType(req.HolidayType); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if err := s.domain.ValidateApplyTo(req.ApplyTo); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if req.Status != nil {
		if err := s.domain.ValidateStatus(*req.Status); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	holiday, err := buildHolidayCreateModel(req, createdBy)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid start_date/end_date; use RFC3339 or YYYY-MM-DD")
	}
	if err := s.domain.ValidateDateRange(holiday.StartDate, holiday.EndDate); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	if err := s.holidayRepo.Create(ctx, holiday, req.StoreIDs); err != nil {
		return nil, err
	}
	return toHolidayResponse(holiday, req.StoreIDs), nil
}

// UpdateHoliday updates an existing holiday by id.
func (s *Service) UpdateHoliday(ctx context.Context, id uint, req *dto.UpdateStoreHolidayRequest) (*dto.StoreHolidayResponse, error) {
	holiday, err := s.holidayRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("holiday not found")
		}
		return nil, err
	}

	if req.Name != nil {
		if err := s.domain.ValidateName(*req.Name); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.HolidayType != nil {
		if err := s.domain.ValidateHolidayType(*req.HolidayType); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.ApplyTo != nil {
		if err := s.domain.ValidateApplyTo(*req.ApplyTo); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.Status != nil {
		if err := s.domain.ValidateStatus(*req.Status); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	if err := applyHolidayUpdate(holiday, req); err != nil {
		return nil, utils.ErrBadRequest("invalid start_date/end_date; use RFC3339 or YYYY-MM-DD")
	}
	if err := s.domain.ValidateDateRange(holiday.StartDate, holiday.EndDate); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	storeIDs := req.StoreIDs
	if storeIDs == nil {
		storeIDs, err = s.holidayRepo.StoreIDsForHoliday(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	if err := s.holidayRepo.Save(ctx, holiday, storeIDs); err != nil {
		return nil, err
	}
	return toHolidayResponse(holiday, storeIDs), nil
}

// GetHoliday loads a single holiday by id.
func (s *Service) GetHoliday(ctx context.Context, id uint) (*dto.StoreHolidayResponse, error) {
	holiday, err := s.holidayRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("holiday not found")
		}
		return nil, err
	}
	return toHolidayResponse(holiday, storeIDsFromModel(holiday)), nil
}

// DeleteHoliday removes a holiday by id.
func (s *Service) DeleteHoliday(ctx context.Context, id uint) error {
	rows, err := s.holidayRepo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrNotFound("holiday not found")
	}
	return nil
}

// ListHolidays returns paginated holidays matching filters.
func (s *Service) ListHolidays(ctx context.Context, req *dto.ListStoreHolidaysRequest) ([]dto.StoreHolidayResponse, int64, error) {
	items, total, err := s.holidayRepo.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.StoreHolidayResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *toHolidayResponse(&items[i], storeIDsFromModel(&items[i])))
	}
	return resp, total, nil
}

// DuplicateHoliday clones a holiday as a new draft (dates unchanged; caller may update after).
func (s *Service) DuplicateHoliday(ctx context.Context, id uint) (*dto.StoreHolidayResponse, error) {
	original, err := s.holidayRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("holiday not found")
		}
		return nil, err
	}
	storeIDs, err := s.holidayRepo.StoreIDsForHoliday(ctx, id)
	if err != nil {
		return nil, err
	}

	clone := &models.StoreHoliday{
		Name:           original.Name + " (copy)",
		Description:    original.Description,
		HolidayType:    original.HolidayType,
		StartDate:      original.StartDate,
		EndDate:        original.EndDate,
		IsRecurring:    original.IsRecurring,
		RecurrenceRule: original.RecurrenceRule,
		ApplyTo:        original.ApplyTo,
		VendorID:       original.VendorID,
		Region:         original.Region,
		Priority:       original.Priority,
		Status:         constants.CalendarStatusDraft,
		Notes:          original.Notes,
		CreatedBy:      original.CreatedBy,
	}
	if err := s.holidayRepo.Create(ctx, clone, storeIDs); err != nil {
		return nil, err
	}
	return toHolidayResponse(clone, storeIDs), nil
}

// PublishHoliday transitions a holiday from draft to published.
func (s *Service) PublishHoliday(ctx context.Context, id uint) (*dto.StoreHolidayResponse, error) {
	if err := s.holidayRepo.SetStatus(ctx, id, constants.CalendarStatusPublished); err != nil {
		return nil, err
	}
	return s.GetHoliday(ctx, id)
}

func storeIDsFromModel(h *models.StoreHoliday) []uint {
	ids := make([]uint, 0, len(h.Stores))
	for _, store := range h.Stores {
		ids = append(ids, store.ID)
	}
	return ids
}
