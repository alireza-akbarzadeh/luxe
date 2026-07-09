package menu

import (
	"context"
	"errors"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// Queries orchestrates menu read use cases.
type Queries struct {
	repo      *postgres.MenuRepository
	adminRepo *postgres.AdminRepository
}

// NewQueries creates menu query use cases.
func NewQueries(repo *postgres.MenuRepository, adminRepo *postgres.AdminRepository) *Queries {
	return &Queries{repo: repo, adminRepo: adminRepo}
}

// ListGroups returns all menu groups.
func (q *Queries) ListGroups(ctx context.Context) ([]models.MenuGroup, error) {
	return q.repo.ListGroups(ctx)
}

// GetGroupByID loads a menu group by id.
func (q *Queries) GetGroupByID(ctx context.Context, id uint) (*models.MenuGroup, error) {
	group, err := q.repo.FindGroupByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return group, err
}

// ListAllItems returns menu items, optionally as a tree.
func (q *Queries) ListAllItems(ctx context.Context, flat bool) ([]models.MenuItem, error) {
	allItems, err := q.repo.ListAllItems(ctx)
	if err != nil {
		return nil, err
	}
	if flat {
		return allItems, nil
	}
	return buildTree(allItems), nil
}

func buildTree(items []models.MenuItem) []models.MenuItem {
	itemMap := make(map[uint]*models.MenuItem)
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}

	var roots []models.MenuItem
	for i := range items {
		parentID := items[i].ParentID
		if parentID == nil || *parentID == 0 {
			roots = append(roots, items[i])
		} else if parent, exists := itemMap[*parentID]; exists {
			parent.Children = append(parent.Children, items[i])
		}
	}
	return roots
}

// GetItemByID loads a menu item by id.
func (q *Queries) GetItemByID(ctx context.Context, id uint) (*models.MenuItem, error) {
	item, err := q.repo.FindItemByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return item, err
}

// GetUserMenu returns filtered sidebar groups for a role and search term.
func (q *Queries) GetUserMenu(ctx context.Context, userRole string, search string) ([]dto.SidebarGroup, error) {
	groups, err := q.repo.ListGroupsWithItems(ctx)
	if err != nil {
		return nil, err
	}

	var allItems []models.MenuItem
	for _, g := range groups {
		allItems = append(allItems, g.Items...)
	}

	itemMap, roots := buildFullTree(allItems)

	var result []dto.SidebarGroup
	for _, group := range groups {
		var groupRoots []models.MenuItem
		for _, root := range roots {
			if root.GroupID == group.ID {
				groupRoots = append(groupRoots, root)
			}
		}
		visibleItems := filterItems(groupRoots, userRole, search, itemMap)
		if len(visibleItems) > 0 {
			result = append(result, dto.SidebarGroup{
				Group: group.Name,
				Items: visibleItems,
			})
		}
	}
	return result, nil
}

func filterItems(items []models.MenuItem, userRole, search string, itemMap map[uint]*models.MenuItem) []dto.SidebarItem {
	var result []dto.SidebarItem
	for _, item := range items {
		if item.Permission != nil && *item.Permission != userRole && userRole != "admin" {
			continue
		}
		filteredChildren := filterItems(item.Children, userRole, search, itemMap)
		matchesSearch := search == "" || matchesSearch(item.Label, item.Href, search)
		if search != "" && !matchesSearch && len(filteredChildren) == 0 {
			continue
		}
		sidebarItem := dto.SidebarItem{
			Label:    item.Label,
			Icon:     item.Icon,
			Children: filteredChildren,
		}
		if item.Href != nil {
			sidebarItem.Href = *item.Href
		}
		result = append(result, sidebarItem)
	}
	return result
}

func buildFullTree(items []models.MenuItem) (map[uint]*models.MenuItem, []models.MenuItem) {
	itemMap := make(map[uint]*models.MenuItem)
	for i := range items {
		itemMap[items[i].ID] = &items[i]
		itemMap[items[i].ID].Children = []models.MenuItem{}
	}
	var roots []models.MenuItem
	for i := range items {
		parentID := items[i].ParentID
		if parentID == nil || *parentID == 0 {
			roots = append(roots, items[i])
		} else if parent, ok := itemMap[*parentID]; ok {
			parent.Children = append(parent.Children, items[i])
		}
	}
	return itemMap, roots
}

func matchesSearch(label string, href *string, search string) bool {
	searchLower := strings.ToLower(search)
	if strings.Contains(strings.ToLower(label), searchLower) {
		return true
	}
	if href != nil && strings.Contains(strings.ToLower(*href), searchLower) {
		return true
	}
	return false
}

// GetUserMenuStructure returns filtered menu groups with nested items.
func (q *Queries) GetUserMenuStructure(ctx context.Context, userRole string, search string) ([]dto.MenuGroupResponse, error) {
	groups, err := q.repo.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	allItems, err := q.repo.ListAllItems(ctx)
	if err != nil {
		return nil, err
	}

	itemsByGroup := make(map[uint][]models.MenuItem)
	for _, item := range allItems {
		itemsByGroup[item.GroupID] = append(itemsByGroup[item.GroupID], item)
	}

	var filterAndBuild func(items []models.MenuItem, parentID *uint) []dto.MenuItemResponse
	filterAndBuild = func(items []models.MenuItem, parentID *uint) []dto.MenuItemResponse {
		var result []dto.MenuItemResponse
		for _, item := range items {
			if (parentID == nil && item.ParentID != nil) ||
				(parentID != nil && (item.ParentID == nil || *item.ParentID != *parentID)) {
				continue
			}
			if item.Permission != nil && *item.Permission != userRole && userRole != "admin" {
				continue
			}
			matches := search == "" || matchesSearch(item.Label, item.Href, search)
			children := filterAndBuild(items, &item.ID)
			if search != "" && !matches && len(children) == 0 {
				continue
			}
			result = append(result, dto.MenuItemResponse{
				ID:           item.ID,
				GroupID:      item.GroupID,
				ParentID:     item.ParentID,
				Label:        item.Label,
				Href:         item.Href,
				Icon:         item.Icon,
				Permission:   item.Permission,
				DisplayOrder: item.DisplayOrder,
				CreatedAt:    item.CreatedAt,
				UpdatedAt:    item.UpdatedAt,
				Children:     children,
			})
		}
		return result
	}

	var response []dto.MenuGroupResponse
	for _, group := range groups {
		nestedItems := filterAndBuild(itemsByGroup[group.ID], nil)
		if len(nestedItems) > 0 {
			response = append(response, dto.MenuGroupResponse{
				ID:           group.ID,
				Name:         group.Name,
				DisplayOrder: group.DisplayOrder,
				CreatedAt:    group.CreatedAt,
				UpdatedAt:    group.UpdatedAt,
				Items:        nestedItems,
			})
		}
	}

	if q.adminRepo != nil {
		q.applyMenuBadges(ctx, response)
	}

	return response, nil
}

func (q *Queries) applyMenuBadges(ctx context.Context, groups []dto.MenuGroupResponse) {
	pendingOrders, err := q.adminRepo.CountPendingOrders(ctx)
	if err != nil {
		return
	}
	lowStock, err := q.adminRepo.CountLowStockProducts(ctx)
	if err != nil {
		return
	}

	var walk func(items []dto.MenuItemResponse)
	walk = func(items []dto.MenuItemResponse) {
		for i := range items {
			if items[i].Href != nil {
				switch *items[i].Href {
				case "/dashboard/orders":
					if pendingOrders > 0 {
						count := pendingOrders
						variant := "warning"
						items[i].BadgeCount = &count
						items[i].BadgeVariant = &variant
					}
				case "/dashboard/inventory":
					if lowStock > 0 {
						count := lowStock
						variant := "destructive"
						items[i].BadgeCount = &count
						items[i].BadgeVariant = &variant
					}
				}
			}
			if len(items[i].Children) > 0 {
				walk(items[i].Children)
			}
		}
	}

	for gi := range groups {
		walk(groups[gi].Items)
	}
}
