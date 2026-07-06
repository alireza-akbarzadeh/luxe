package reversemarketplace

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service serves reverse marketplace buyer and vendor flows.
type Service struct {
	repo *postgres.ReverseMarketplaceRepository
}

// NewService wires reverse marketplace use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: postgres.NewReverseMarketplaceRepository(db)}
}

// ListOpen returns open buyer requests for storefront discovery.
func (s *Service) ListOpen(ctx context.Context, limit, offset int) (dto.ReverseMarketplaceRequestListResponse, error) {
	requests, total, err := s.repo.ListOpen(ctx, limit, offset)
	if err != nil {
		return dto.ReverseMarketplaceRequestListResponse{}, err
	}

	items := make([]dto.ReverseMarketplaceRequestListItem, 0, len(requests))
	for i := range requests {
		items = append(items, toListItem(&requests[i]))
	}

	return dto.ReverseMarketplaceRequestListResponse{
		Requests: items,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// ListForUser returns requests created by the authenticated buyer.
func (s *Service) ListForUser(ctx context.Context, userID uint, limit, offset int) (dto.ReverseMarketplaceRequestListResponse, error) {
	requests, total, err := s.repo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return dto.ReverseMarketplaceRequestListResponse{}, err
	}

	items := make([]dto.ReverseMarketplaceRequestListItem, 0, len(requests))
	for i := range requests {
		items = append(items, toListItem(&requests[i]))
	}

	return dto.ReverseMarketplaceRequestListResponse{
		Requests: items,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// GetByID returns a request with vendor offers.
func (s *Service) GetByID(ctx context.Context, id uint) (dto.ReverseMarketplaceRequestResponse, error) {
	request, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ReverseMarketplaceRequestResponse{}, utils.ErrNotFound("request not found")
		}
		return dto.ReverseMarketplaceRequestResponse{}, err
	}
	return toResponse(request), nil
}

// CreateRequest posts a new wanted listing for the authenticated buyer.
func (s *Service) CreateRequest(ctx context.Context, userID uint, req dto.CreateReverseMarketplaceRequest) (dto.ReverseMarketplaceRequestResponse, error) {
	if req.BudgetMin != nil && req.BudgetMax != nil && *req.BudgetMin > *req.BudgetMax {
		return dto.ReverseMarketplaceRequestResponse{}, utils.ErrBadRequest("budget_min cannot exceed budget_max")
	}

	request := &models.ReverseMarketplaceRequest{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		BudgetMin:   req.BudgetMin,
		BudgetMax:   req.BudgetMax,
		Status:      constants.ReverseMarketplaceRequestStatusOpen,
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return dto.ReverseMarketplaceRequestResponse{}, err
	}
	return toResponse(request), nil
}

// SubmitOffer lets a vendor respond to an open buyer request.
func (s *Service) SubmitOffer(
	ctx context.Context,
	storeID, vendorUserID, requestID uint,
	req dto.CreateReverseMarketplaceOfferRequest,
) (dto.ReverseMarketplaceOfferResponse, error) {
	request, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ReverseMarketplaceOfferResponse{}, utils.ErrNotFound("request not found")
		}
		return dto.ReverseMarketplaceOfferResponse{}, err
	}

	if request.Status != constants.ReverseMarketplaceRequestStatusOpen {
		return dto.ReverseMarketplaceOfferResponse{}, utils.ErrBadRequest("request is not open for offers")
	}

	if _, err := s.repo.GetOfferByStoreAndRequest(ctx, storeID, requestID); err == nil {
		return dto.ReverseMarketplaceOfferResponse{}, utils.ErrBadRequest("your store already submitted an offer for this request")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.ReverseMarketplaceOfferResponse{}, err
	}

	offer := &models.ReverseMarketplaceOffer{
		RequestID:    requestID,
		StoreID:      storeID,
		VendorUserID: vendorUserID,
		Message:      req.Message,
		OfferedPrice: req.OfferedPrice,
		Status:       constants.ReverseMarketplaceOfferStatusPending,
	}

	if err := s.repo.CreateOffer(ctx, offer); err != nil {
		if postgres.IsUniqueViolation(err) {
			return dto.ReverseMarketplaceOfferResponse{}, utils.ErrBadRequest("your store already submitted an offer for this request")
		}
		return dto.ReverseMarketplaceOfferResponse{}, err
	}

	return dto.ReverseMarketplaceOfferResponse{
		ID:           offer.ID,
		StoreID:      offer.StoreID,
		Message:      offer.Message,
		OfferedPrice: offer.OfferedPrice,
		Status:       offer.Status,
		CreatedAt:    offer.CreatedAt.Format(time.RFC3339),
	}, nil
}

func toListItem(request *models.ReverseMarketplaceRequest) dto.ReverseMarketplaceRequestListItem {
	if request == nil {
		return dto.ReverseMarketplaceRequestListItem{}
	}
	return dto.ReverseMarketplaceRequestListItem{
		ID:          request.ID,
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		BudgetMin:   request.BudgetMin,
		BudgetMax:   request.BudgetMax,
		Status:      request.Status,
		OfferCount:  len(request.Offers),
		CreatedAt:   request.CreatedAt.Format(time.RFC3339),
	}
}

func toResponse(request *models.ReverseMarketplaceRequest) dto.ReverseMarketplaceRequestResponse {
	if request == nil {
		return dto.ReverseMarketplaceRequestResponse{}
	}

	offers := make([]dto.ReverseMarketplaceOfferResponse, 0, len(request.Offers))
	for _, offer := range request.Offers {
		storeName := ""
		if offer.Store != nil {
			storeName = offer.Store.Name
		}
		offers = append(offers, dto.ReverseMarketplaceOfferResponse{
			ID:           offer.ID,
			StoreID:      offer.StoreID,
			StoreName:    storeName,
			Message:      offer.Message,
			OfferedPrice: offer.OfferedPrice,
			Status:       offer.Status,
			CreatedAt:    offer.CreatedAt.Format(time.RFC3339),
		})
	}

	return dto.ReverseMarketplaceRequestResponse{
		ID:          request.ID,
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		BudgetMin:   request.BudgetMin,
		BudgetMax:   request.BudgetMax,
		Status:      request.Status,
		CreatedAt:   request.CreatedAt.Format(time.RFC3339),
		Offers:      offers,
	}
}
