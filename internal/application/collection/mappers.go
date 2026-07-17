package collection

import (
	"encoding/json"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/datatypes"
)

func parseScheduleTime(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := *raw
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeMode(mode string, collectionType string) string {
	if mode != "" {
		return mode
	}
	switch collectionType {
	case "manual":
		return "manual"
	case "hybrid":
		return "hybrid"
	default:
		return "dynamic"
	}
}

func legacyCollectionType(mode string, collectionType string) string {
	if collectionType == "manual" || collectionType == "smart" {
		return collectionType
	}
	if mode == "manual" {
		return "manual"
	}
	return "smart"
}

func marshalRules(rules *dto.CollectionRules) (datatypes.JSON, error) {
	if rules == nil {
		return datatypes.JSON([]byte(`{}`)), nil
	}
	encoded, err := json.Marshal(rules)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(encoded), nil
}

func unmarshalRules(raw datatypes.JSON) *dto.CollectionRules {
	if len(raw) == 0 || string(raw) == "" || string(raw) == "{}" {
		return nil
	}
	var rules dto.CollectionRules
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil
	}
	if rules.Operator == "" && len(rules.Conditions) == 0 && len(rules.Groups) == 0 {
		return nil
	}
	if rules.Operator == "" {
		rules.Operator = "and"
	}
	normalizeLegacyConditionOperators(&rules)
	return &rules
}

// normalizeLegacyConditionOperators maps legacy "op" aliases if present via re-marshal path.
func normalizeLegacyConditionOperators(rules *dto.CollectionRules) {
	for i := range rules.Conditions {
		if rules.Conditions[i].Operator == "" {
			rules.Conditions[i].Operator = "eq"
		}
	}
	for gi := range rules.Groups {
		normalizeGroupOperators(&rules.Groups[gi])
	}
}

func normalizeGroupOperators(group *dto.CollectionRuleGroup) {
	if group.Operator == "" {
		group.Operator = "and"
	}
	for i := range group.Conditions {
		if group.Conditions[i].Operator == "" {
			group.Conditions[i].Operator = "eq"
		}
	}
	for i := range group.Groups {
		normalizeGroupOperators(&group.Groups[i])
	}
}

// BuildCreateModel maps a create DTO to a collection model.
func BuildCreateModel(req *dto.CreateCollectionRequest, slug string) (*models.Collection, error) {
	status := req.Status
	if status == "" {
		status = "draft"
	}
	mode := normalizeMode(req.Mode, req.CollectionType)
	collectionType := legacyCollectionType(mode, req.CollectionType)
	href := req.Href
	if href == "" {
		href = "/shop"
	}
	cta := req.CTALabel
	if cta == "" {
		cta = "Shop collection"
	}
	startsAt, err := parseScheduleTime(req.StartsAt)
	if err != nil {
		return nil, err
	}
	endsAt, err := parseScheduleTime(req.EndsAt)
	if err != nil {
		return nil, err
	}
	if err := ValidateCollectionRules(mode, req.Rules); err != nil {
		return nil, err
	}
	rulesJSON, err := marshalRules(req.Rules)
	if err != nil {
		return nil, err
	}
	status = normalizePublishStatus(status, startsAt)
	isIndexable := true
	if req.IsIndexable != nil {
		isIndexable = *req.IsIndexable
	}
	overlayOpacity := 0.25
	if req.OverlayOpacity != nil {
		overlayOpacity = *req.OverlayOpacity
	}
	return &models.Collection{
		Slug:              slug,
		Eyebrow:           req.Eyebrow,
		Title:             req.Title,
		Subtitle:          req.Subtitle,
		Description:       req.Description,
		Href:              href,
		ImageURL:          req.ImageURL,
		CTALabel:          cta,
		SortOrder:         req.SortOrder,
		Status:            status,
		CollectionType:    collectionType,
		Mode:              mode,
		StartsAt:          startsAt,
		EndsAt:            endsAt,
		PreviewSort:       req.PreviewSort,
		PreviewIsNew:      req.PreviewIsNew,
		PreviewCategoryID: req.PreviewCategoryID,
		Theme:             req.Theme,
		SortKey:           req.SortKey,
		RulesJSON:         rulesJSON,
		SEOTitle:          req.SEOTitle,
		SEODescription:    req.SEODescription,
		MetaKeywords:      req.MetaKeywords,
		OGTitle:           req.OGTitle,
		OGDescription:     req.OGDescription,
		OGImageURL:        req.OGImageURL,
		TwitterTitle:      req.TwitterTitle,
		TwitterDescription: req.TwitterDescription,
		TwitterImageURL:   req.TwitterImageURL,
		CanonicalURL:      req.CanonicalURL,
		RobotsDirectives:  req.RobotsDirectives,
		IsIndexable:       isIndexable,
		HeroTitle:         req.HeroTitle,
		HeroDescription:   req.HeroDescription,
		DesktopImageURL:   req.DesktopImageURL,
		TabletImageURL:    req.TabletImageURL,
		MobileImageURL:    req.MobileImageURL,
		OverlayOpacity:    overlayOpacity,
		ThemeVariant:      req.ThemeVariant,
	}, nil
}

// ApplyUpdateDTO mutates a loaded collection from an update DTO.
func ApplyUpdateDTO(collection *models.Collection, req *dto.UpdateCollectionRequest, newSlug string) error {
	if newSlug != "" {
		collection.Slug = newSlug
	}
	if req.Eyebrow != nil {
		collection.Eyebrow = *req.Eyebrow
	}
	if req.Title != nil {
		collection.Title = *req.Title
	}
	if req.Subtitle != nil {
		collection.Subtitle = *req.Subtitle
	}
	if req.Description != nil {
		collection.Description = *req.Description
	}
	if req.Href != nil {
		collection.Href = *req.Href
	}
	if req.ImageURL != nil {
		collection.ImageURL = *req.ImageURL
	}
	if req.CTALabel != nil {
		collection.CTALabel = *req.CTALabel
	}
	if req.SortOrder != nil {
		collection.SortOrder = *req.SortOrder
	}
	if req.Status != nil {
		collection.Status = *req.Status
	}
	if req.CollectionType != nil {
		collection.CollectionType = *req.CollectionType
		if req.Mode == nil {
			collection.Mode = normalizeMode("", *req.CollectionType)
		}
	}
	if req.Mode != nil {
		collection.Mode = *req.Mode
		collection.CollectionType = legacyCollectionType(*req.Mode, collection.CollectionType)
	}
	if req.StartsAt != nil {
		startsAt, err := parseScheduleTime(req.StartsAt)
		if err != nil {
			return err
		}
		collection.StartsAt = startsAt
	}
	if req.EndsAt != nil {
		endsAt, err := parseScheduleTime(req.EndsAt)
		if err != nil {
			return err
		}
		collection.EndsAt = endsAt
	}
	if req.Rules != nil {
		mode := collection.Mode
		if req.Mode != nil {
			mode = *req.Mode
		}
		if err := ValidateCollectionRules(mode, req.Rules); err != nil {
			return err
		}
		rulesJSON, err := marshalRules(req.Rules)
		if err != nil {
			return err
		}
		collection.RulesJSON = rulesJSON
	} else if req.Mode != nil {
		mode := *req.Mode
		if err := ValidateCollectionRules(mode, unmarshalRules(collection.RulesJSON)); err != nil {
			return err
		}
	}
	collection.Status = normalizePublishStatus(collection.Status, collection.StartsAt)
	if req.PreviewSort != nil {
		collection.PreviewSort = *req.PreviewSort
	}
	if req.PreviewIsNew != nil {
		collection.PreviewIsNew = req.PreviewIsNew
	}
	if req.PreviewCategoryID != nil {
		collection.PreviewCategoryID = req.PreviewCategoryID
	}
	if req.Theme != nil {
		collection.Theme = *req.Theme
	}
	if req.SortKey != nil {
		collection.SortKey = *req.SortKey
	}
	if req.SEOTitle != nil {
		collection.SEOTitle = *req.SEOTitle
	}
	if req.SEODescription != nil {
		collection.SEODescription = *req.SEODescription
	}
	if req.MetaKeywords != nil {
		collection.MetaKeywords = *req.MetaKeywords
	}
	if req.OGTitle != nil {
		collection.OGTitle = *req.OGTitle
	}
	if req.OGDescription != nil {
		collection.OGDescription = *req.OGDescription
	}
	if req.OGImageURL != nil {
		collection.OGImageURL = *req.OGImageURL
	}
	if req.TwitterTitle != nil {
		collection.TwitterTitle = *req.TwitterTitle
	}
	if req.TwitterDescription != nil {
		collection.TwitterDescription = *req.TwitterDescription
	}
	if req.TwitterImageURL != nil {
		collection.TwitterImageURL = *req.TwitterImageURL
	}
	if req.CanonicalURL != nil {
		collection.CanonicalURL = *req.CanonicalURL
	}
	if req.RobotsDirectives != nil {
		collection.RobotsDirectives = *req.RobotsDirectives
	}
	if req.IsIndexable != nil {
		collection.IsIndexable = *req.IsIndexable
	}
	if req.HeroTitle != nil {
		collection.HeroTitle = *req.HeroTitle
	}
	if req.HeroDescription != nil {
		collection.HeroDescription = *req.HeroDescription
	}
	if req.DesktopImageURL != nil {
		collection.DesktopImageURL = *req.DesktopImageURL
	}
	if req.TabletImageURL != nil {
		collection.TabletImageURL = *req.TabletImageURL
	}
	if req.MobileImageURL != nil {
		collection.MobileImageURL = *req.MobileImageURL
	}
	if req.OverlayOpacity != nil {
		collection.OverlayOpacity = *req.OverlayOpacity
	}
	if req.ThemeVariant != nil {
		collection.ThemeVariant = *req.ThemeVariant
	}
	return nil
}

// ProductIDsFromModel returns ordered product ids from preloaded collection products.
func ProductIDsFromModel(c *models.Collection) []uint {
	if len(c.Products) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(c.Products))
	for _, item := range c.Products {
		ids = append(ids, item.ProductID)
	}
	return ids
}

// ToResponse maps a collection model to an API response.
func ToResponse(c *models.Collection) *dto.CollectionResponse {
	collectionType := c.CollectionType
	if collectionType == "" {
		collectionType = "smart"
	}
	mode := c.Mode
	if mode == "" {
		mode = normalizeMode("", collectionType)
	}
	resp := &dto.CollectionResponse{
		ID:                c.ID,
		Slug:              c.Slug,
		Eyebrow:           c.Eyebrow,
		Title:             c.Title,
		Subtitle:          c.Subtitle,
		Description:       c.Description,
		Href:              c.Href,
		ImageURL:          c.ImageURL,
		CTALabel:          c.CTALabel,
		SortOrder:         c.SortOrder,
		Status:            c.Status,
		CollectionType:    collectionType,
		Mode:              mode,
		StartsAt:          c.StartsAt,
		EndsAt:            c.EndsAt,
		ProductIDs:        ProductIDsFromModel(c),
		ProductOverrides:  ProductOverridesFromModel(c),
		PreviewSort:       c.PreviewSort,
		PreviewIsNew:      c.PreviewIsNew,
		PreviewCategoryID: c.PreviewCategoryID,
		Theme:             c.Theme,
		SortKey:           c.SortKey,
		Rules:             unmarshalRules(c.RulesJSON),
		SEOTitle:          c.SEOTitle,
		SEODescription:    c.SEODescription,
		MetaKeywords:      c.MetaKeywords,
		OGTitle:           c.OGTitle,
		OGDescription:     c.OGDescription,
		OGImageURL:        c.OGImageURL,
		TwitterTitle:      c.TwitterTitle,
		TwitterDescription: c.TwitterDescription,
		TwitterImageURL:   c.TwitterImageURL,
		CanonicalURL:      c.CanonicalURL,
		RobotsDirectives:  c.RobotsDirectives,
		IsIndexable:       c.IsIndexable,
		HeroTitle:         c.HeroTitle,
		HeroDescription:   c.HeroDescription,
		DesktopImageURL:   c.DesktopImageURL,
		TabletImageURL:    c.TabletImageURL,
		MobileImageURL:    c.MobileImageURL,
		OverlayOpacity:    c.OverlayOpacity,
		ThemeVariant:      c.ThemeVariant,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
	if c.WorkflowState != nil {
		resp.WorkflowState = dto.ToStateView(c.WorkflowState)
	}
	return resp
}

// ProductOverridesFromModel returns merchandising overrides from preloaded collection products.
func ProductOverridesFromModel(c *models.Collection) []dto.CollectionProductOverrideInput {
	if len(c.Products) == 0 {
		return nil
	}
	items := make([]dto.CollectionProductOverrideInput, 0, len(c.Products))
	for _, item := range c.Products {
		items = append(items, dto.CollectionProductOverrideInput{
			ProductID:  item.ProductID,
			Position:   item.Position,
			IsPinned:   item.IsPinned,
			IsHidden:   item.IsHidden,
			BoostScore: item.BoostScore,
		})
	}
	return items
}
