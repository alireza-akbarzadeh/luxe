package promotion

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	appcategory "github.com/alireza-akbarzadeh/luxe/internal/application/category"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service handles admin merchandising promotions (flash deals, banners, campaigns).
type Service struct {
	repo *postgres.PromotionRepository
}

// NewService wires the promotion application service.
func NewService(repo *postgres.PromotionRepository) *Service {
	return &Service{repo: repo}
}

func normalizeFlashStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "draft", "ended":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "active"
	}
}

func normalizeSectionStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "published", "archived":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "draft"
	}
}

func normalizeCampaignStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "scheduled", "active", "ended", "archived":
		return strings.TrimSpace(strings.ToLower(status))
	default:
		return "draft"
	}
}

func placementsFromRequest(req *dto.CampaignPlacementsRequest) models.CampaignPlacements {
	if req == nil {
		return models.CampaignPlacements{}
	}
	return models.CampaignPlacements{
		FlashDealIDs:  req.FlashDealIDs,
		SectionIDs:    req.SectionIDs,
		CollectionIDs: req.CollectionIDs,
	}
}

func filtersFromMap(filters map[string]interface{}) datatypes.JSON {
	if len(filters) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	b, err := json.Marshal(filters)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(b)
}

// GetKPIs returns promotion hub summary counts.
func (s *Service) GetKPIs(ctx context.Context) (*dto.PromotionsKPIData, error) {
	now := time.Now()
	flash, banners, campaigns, err := s.repo.CountActivePromotions(ctx, now)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	var scheduled int64
	_, scheduled, err = s.repo.ListCampaigns(ctx, dto.AdminCampaignListFilters{Status: "scheduled", Limit: 1})
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.PromotionsKPIData{
		ActiveFlashDeals:   flash,
		PublishedBanners:   banners,
		ActiveCampaigns:    campaigns,
		ScheduledCampaigns: scheduled,
	}, nil
}

// --- Flash deals ---

func (s *Service) ListFlashDeals(ctx context.Context, filters dto.AdminFlashDealListFilters) ([]models.FlashDeal, int64, error) {
	rows, total, err := s.repo.ListFlashDeals(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return rows, total, nil
}

func (s *Service) GetFlashDeal(ctx context.Context, id uint) (*models.FlashDeal, error) {
	row, err := s.repo.FindFlashDealByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("flash deal not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

func (s *Service) CreateFlashDeal(ctx context.Context, req dto.CreateFlashDealRequest) (*models.FlashDeal, error) {
	row := &models.FlashDeal{
		ProductID:     req.ProductID,
		Title:         strings.TrimSpace(req.Title),
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
		QuantityLimit: req.QuantityLimit,
		SortOrder:     req.SortOrder,
		Status:        normalizeFlashStatus(req.Status),
	}
	if err := s.repo.CreateFlashDeal(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetFlashDeal(ctx, row.ID)
}

func (s *Service) UpdateFlashDeal(ctx context.Context, id uint, req dto.UpdateFlashDealRequest) (*models.FlashDeal, error) {
	row, err := s.GetFlashDeal(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.ProductID != nil {
		row.ProductID = *req.ProductID
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.StartsAt != nil {
		row.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		row.EndsAt = *req.EndsAt
	}
	if req.QuantityLimit != nil {
		row.QuantityLimit = req.QuantityLimit
	}
	if req.SortOrder != nil {
		row.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		row.Status = normalizeFlashStatus(*req.Status)
	}
	if err := s.repo.UpdateFlashDeal(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetFlashDeal(ctx, id)
}

func (s *Service) DeleteFlashDeal(ctx context.Context, id uint) error {
	if _, err := s.GetFlashDeal(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteFlashDeal(ctx, id); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// --- Homepage sections ---

func (s *Service) ListHomepageSections(ctx context.Context, filters dto.AdminHomepageSectionListFilters) ([]models.HomepageSection, int64, error) {
	rows, total, err := s.repo.ListHomepageSections(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return rows, total, nil
}

func (s *Service) GetHomepageSection(ctx context.Context, id uint) (*models.HomepageSection, error) {
	row, err := s.repo.FindHomepageSectionByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("homepage section not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

func (s *Service) CreateHomepageSection(ctx context.Context, req dto.CreateHomepageSectionRequest) (*models.HomepageSection, error) {
	key := strings.TrimSpace(strings.ToLower(req.SectionKey))
	exists, err := s.repo.FindHomepageSectionByKey(ctx, key, 0)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if exists {
		return nil, utils.ErrConflict("section key already exists")
	}

	row := &models.HomepageSection{
		SectionKey: key,
		Title:      strings.TrimSpace(req.Title),
		Href:       strings.TrimSpace(req.Href),
		ImageURL:   strings.TrimSpace(req.ImageURL),
		SortOrder:  req.SortOrder,
		Status:     normalizeSectionStatus(req.Status),
		Filters:    filtersFromMap(req.Filters),
	}
	if err := s.repo.CreateHomepageSection(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetHomepageSection(ctx, row.ID)
}

func (s *Service) UpdateHomepageSection(ctx context.Context, id uint, req dto.UpdateHomepageSectionRequest) (*models.HomepageSection, error) {
	row, err := s.GetHomepageSection(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.SectionKey != nil {
		key := strings.TrimSpace(strings.ToLower(*req.SectionKey))
		exists, err := s.repo.FindHomepageSectionByKey(ctx, key, id)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		if exists {
			return nil, utils.ErrConflict("section key already exists")
		}
		row.SectionKey = key
	}
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.Href != nil {
		row.Href = strings.TrimSpace(*req.Href)
	}
	if req.ImageURL != nil {
		row.ImageURL = strings.TrimSpace(*req.ImageURL)
	}
	if req.SortOrder != nil {
		row.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		row.Status = normalizeSectionStatus(*req.Status)
	}
	if req.Filters != nil {
		row.Filters = filtersFromMap(req.Filters)
	}
	if err := s.repo.UpdateHomepageSection(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetHomepageSection(ctx, id)
}

func (s *Service) DeleteHomepageSection(ctx context.Context, id uint) error {
	if _, err := s.GetHomepageSection(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteHomepageSection(ctx, id); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// --- Campaigns ---

func (s *Service) ListCampaigns(ctx context.Context, filters dto.AdminCampaignListFilters) ([]models.Campaign, int64, error) {
	rows, total, err := s.repo.ListCampaigns(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return rows, total, nil
}

func (s *Service) GetCampaign(ctx context.Context, id uint) (*models.Campaign, error) {
	row, err := s.repo.FindCampaignByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("campaign not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return row, nil
}

func (s *Service) CreateCampaign(ctx context.Context, req dto.CreateCampaignRequest) (*models.Campaign, error) {
	slug := strings.TrimSpace(strings.ToLower(req.Slug))
	if slug == "" {
		slug = appcategory.GenerateSlug(req.Name)
	}
	exists, err := s.repo.FindCampaignBySlug(ctx, slug, 0)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if exists {
		return nil, utils.ErrConflict("campaign slug already exists")
	}

	row := &models.Campaign{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: strings.TrimSpace(req.Description),
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Status:      normalizeCampaignStatus(req.Status),
		Placements:  models.MarshalCampaignPlacements(placementsFromRequest(req.Placements)),
	}
	if err := s.repo.CreateCampaign(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetCampaign(ctx, row.ID)
}

func (s *Service) UpdateCampaign(ctx context.Context, id uint, req dto.UpdateCampaignRequest) (*models.Campaign, error) {
	row, err := s.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		row.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil && strings.TrimSpace(*req.Slug) != "" {
		slug := strings.TrimSpace(strings.ToLower(*req.Slug))
		exists, err := s.repo.FindCampaignBySlug(ctx, slug, id)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		if exists {
			return nil, utils.ErrConflict("campaign slug already exists")
		}
		row.Slug = slug
	}
	if req.Description != nil {
		row.Description = strings.TrimSpace(*req.Description)
	}
	if req.StartsAt != nil {
		row.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		row.EndsAt = req.EndsAt
	}
	if req.Status != nil {
		row.Status = normalizeCampaignStatus(*req.Status)
	}
	if req.Placements != nil {
		row.Placements = models.MarshalCampaignPlacements(placementsFromRequest(req.Placements))
	}
	if err := s.repo.UpdateCampaign(ctx, row); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return s.GetCampaign(ctx, id)
}

func (s *Service) DeleteCampaign(ctx context.Context, id uint) error {
	if _, err := s.GetCampaign(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteCampaign(ctx, id); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}
