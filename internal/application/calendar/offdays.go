package calendar

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// CreateOffDay creates a vendor-initiated off day after domain validation.
func (s *Service) CreateOffDay(ctx context.Context, req *dto.CreateVendorOffDayRequest) (*dto.VendorOffDayResponse, error) {
	if err := s.domain.ValidateOffType(req.OffType); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}
	if req.Status != nil {
		if err := s.domain.ValidateStatus(*req.Status); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	offDay, err := buildOffDayCreateModel(req)
	if err != nil {
		return nil, utils.ErrBadRequest("invalid start_date/end_date; use RFC3339 or YYYY-MM-DD")
	}
	if err := s.domain.ValidateDateRange(offDay.StartDate, offDay.EndDate); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	if err := s.offDayRepo.Create(ctx, offDay); err != nil {
		return nil, err
	}
	return toOffDayResponse(offDay), nil
}

// UpdateOffDay updates an existing vendor off day by id.
func (s *Service) UpdateOffDay(ctx context.Context, id uint, req *dto.UpdateVendorOffDayRequest) (*dto.VendorOffDayResponse, error) {
	offDay, err := s.offDayRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("vendor off day not found")
		}
		return nil, err
	}

	if req.OffType != nil {
		if err := s.domain.ValidateOffType(*req.OffType); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}
	if req.Status != nil {
		if err := s.domain.ValidateStatus(*req.Status); err != nil {
			return nil, utils.ErrBadRequest(err.Error())
		}
	}

	if err := applyOffDayUpdate(offDay, req); err != nil {
		return nil, utils.ErrBadRequest("invalid start_date/end_date; use RFC3339 or YYYY-MM-DD")
	}
	if err := s.domain.ValidateDateRange(offDay.StartDate, offDay.EndDate); err != nil {
		return nil, utils.ErrBadRequest(err.Error())
	}

	if err := s.offDayRepo.Save(ctx, offDay); err != nil {
		return nil, err
	}
	return toOffDayResponse(offDay), nil
}

// DeleteOffDay removes a vendor off day by id.
func (s *Service) DeleteOffDay(ctx context.Context, id uint) error {
	rows, err := s.offDayRepo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrNotFound("vendor off day not found")
	}
	return nil
}

// ListOffDays returns paginated vendor off days matching filters.
func (s *Service) ListOffDays(ctx context.Context, req *dto.ListVendorOffDaysRequest) ([]dto.VendorOffDayResponse, int64, error) {
	items, total, err := s.offDayRepo.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.VendorOffDayResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *toOffDayResponse(&items[i]))
	}
	return resp, total, nil
}
