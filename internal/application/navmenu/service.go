package navmenu

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// Queries orchestrates nav menu read use cases.
type Queries struct {
	repo *postgres.NavMenuRepository
}

// NewQueries creates nav menu query use cases.
func NewQueries(repo *postgres.NavMenuRepository) *Queries {
	return &Queries{repo: repo}
}

// GetAll returns all nav menu items as DTOs.
func (q *Queries) GetAll(ctx context.Context) ([]dto.NavItemResponse, error) {
	menus, err := q.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.NavItemResponse, 0, len(menus))
	for i := range menus {
		resp, err := dto.ToNavItemResponse(ctx, &menus[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *resp)
	}
	return result, nil
}

// GetByID returns a nav menu item by id.
func (q *Queries) GetByID(ctx context.Context, id uint) (*dto.NavItemResponse, error) {
	menu, err := q.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(ctx, menu)
}

// Commands orchestrates nav menu write use cases.
type Commands struct {
	repo *postgres.NavMenuRepository
}

// NewCommands creates nav menu command use cases.
func NewCommands(repo *postgres.NavMenuRepository) *Commands {
	return &Commands{repo: repo}
}

// Create inserts a nav menu item.
func (c *Commands) Create(ctx context.Context, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	viewAllJSON, columnsJSON, featuredJSON := dto.NavMenuJSONFromRequest(req)

	menu := models.NavMenu{
		Label:     req.Label,
		LabelI18n: dto.LabelI18nFromRequest(req, nil),
		Type:      req.Type,
		Href:      req.Href,
		Badge:     req.Badge,
		BadgeI18n: dto.BadgeI18nFromRequest(req, nil),
		ViewAll:   viewAllJSON,
		Columns:   columnsJSON,
		Featured:  featuredJSON,
		SortOrder: req.Order,
	}
	if err := c.repo.Create(ctx, &menu); err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(ctx, &menu)
}

// Update modifies a nav menu item.
func (c *Commands) Update(ctx context.Context, id uint, req *dto.UpsertNavMenuRequest) (*dto.NavItemResponse, error) {
	menu, err := c.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	viewAllJSON, columnsJSON, featuredJSON := dto.NavMenuJSONFromRequest(req)

	menu.Label = req.Label
	menu.LabelI18n = dto.LabelI18nFromRequest(req, menu.LabelI18n)
	menu.Type = req.Type
	menu.Href = req.Href
	menu.Badge = req.Badge
	menu.BadgeI18n = dto.BadgeI18nFromRequest(req, menu.BadgeI18n)
	menu.ViewAll = viewAllJSON
	menu.Columns = columnsJSON
	menu.Featured = featuredJSON
	menu.SortOrder = req.Order

	if err := c.repo.Save(ctx, menu); err != nil {
		return nil, err
	}
	return dto.ToNavItemResponse(ctx, menu)
}

// Reorder updates sort order for multiple nav menus.
func (c *Commands) Reorder(ctx context.Context, req *dto.ReorderNavMenusRequest) error {
	items := make([]struct {
		ID    uint
		Order int
	}, len(req.Items))
	for i, item := range req.Items {
		items[i].ID = item.ID
		items[i].Order = item.Order
	}
	return c.repo.Reorder(ctx, items)
}

// Delete removes a nav menu by id.
func (c *Commands) Delete(ctx context.Context, id uint) error {
	rows, err := c.repo.DeleteByID(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
