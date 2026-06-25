package menu

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Commands orchestrates menu write use cases.
type Commands struct {
	repo *postgres.MenuRepository
}

// NewCommands creates menu command use cases.
func NewCommands(repo *postgres.MenuRepository) *Commands {
	return &Commands{repo: repo}
}

// CreateGroup inserts a menu group.
func (c *Commands) CreateGroup(ctx context.Context, req *dto.CreateMenuGroupRequest) (*models.MenuGroup, error) {
	group := &models.MenuGroup{
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
	}
	if err := c.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// UpdateGroup updates a menu group.
func (c *Commands) UpdateGroup(ctx context.Context, id uint, req *dto.UpdateMenuGroupRequest) (*models.MenuGroup, error) {
	group, err := c.repo.FindGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("group not found")
		}
		return nil, err
	}
	group.Name = req.Name
	group.DisplayOrder = req.DisplayOrder
	if err := c.repo.SaveGroup(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// DeleteGroup removes a menu group.
func (c *Commands) DeleteGroup(ctx context.Context, id uint) error {
	return c.repo.DeleteGroup(ctx, id)
}

// CreateItem inserts a menu item after validating group and parent.
func (c *Commands) CreateItem(ctx context.Context, req *dto.CreateMenuItemRequest) (*models.MenuItem, error) {
	if err := c.repo.FindGroupExists(ctx, req.GroupID); err != nil {
		return nil, errors.New("invalid group_id")
	}
	if req.ParentID != nil && *req.ParentID != 0 {
		parent, err := c.repo.FindItemExists(ctx, *req.ParentID)
		if err != nil {
			return nil, utils.ErrBadRequest("invalid parent_id")
		}
		if parent.GroupID != req.GroupID {
			return nil, utils.ErrBadRequest("parent_id must belong to the same group")
		}
	}
	item := &models.MenuItem{
		GroupID:      req.GroupID,
		ParentID:     req.ParentID,
		Label:        req.Label,
		Href:         req.Href,
		Icon:         req.Icon,
		Permission:   req.Permission,
		DisplayOrder: req.DisplayOrder,
	}
	if err := c.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateItem updates a menu item.
func (c *Commands) UpdateItem(ctx context.Context, id uint, req *dto.UpdateMenuItemRequest) (*models.MenuItem, error) {
	item, err := c.repo.FindItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("item not found")
		}
		return nil, err
	}
	if item.GroupID != req.GroupID {
		if err := c.repo.FindGroupExists(ctx, req.GroupID); err != nil {
			return nil, errors.New("invalid group_id")
		}
		item.GroupID = req.GroupID
	}
	if req.ParentID != nil && *req.ParentID != 0 {
		parent, err := c.repo.FindItemExists(ctx, *req.ParentID)
		if err != nil {
			return nil, errors.New("invalid parent_id")
		}
		if parent.GroupID != req.GroupID {
			return nil, errors.New("parent_id must belong to same group")
		}
		item.ParentID = req.ParentID
	} else {
		item.ParentID = nil
	}
	item.Label = req.Label
	item.Href = req.Href
	item.Icon = req.Icon
	item.Permission = req.Permission
	item.DisplayOrder = req.DisplayOrder
	if err := c.repo.SaveItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteItem removes a menu item.
func (c *Commands) DeleteItem(ctx context.Context, id uint) error {
	return c.repo.DeleteItem(ctx, id)
}
