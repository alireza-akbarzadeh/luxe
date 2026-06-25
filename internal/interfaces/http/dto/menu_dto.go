package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ------------------- Request DTOs -------------------

type CreateMenuGroupRequest struct {
	Name         string `json:"name" binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type UpdateMenuGroupRequest struct {
	Name         string `json:"name" binding:"required"`
	DisplayOrder int    `json:"display_order"`
}

type CreateMenuItemRequest struct {
	GroupID      uint    `json:"group_id" binding:"required"`
	ParentID     *uint   `json:"parent_id"`
	Label        string  `json:"label" binding:"required"`
	Href         *string `json:"href"`
	Icon         string  `json:"icon" binding:"required"`
	Permission   *string `json:"permission"`
	DisplayOrder int     `json:"display_order"`
}

type UpdateMenuItemRequest struct {
	GroupID      uint    `json:"group_id" binding:"required"`
	ParentID     *uint   `json:"parent_id"`
	Label        string  `json:"label" binding:"required"`
	Href         *string `json:"href"`
	Icon         string  `json:"icon" binding:"required"`
	Permission   *string `json:"permission"`
	DisplayOrder int     `json:"display_order"`
}

// ------------------- Response DTOs -------------------

type MenuGroupResponse struct {
	ID           uint               `json:"id"`
	Name         string             `json:"name"`
	DisplayOrder int                `json:"display_order"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Items        []MenuItemResponse `json:"items,omitempty"`
}

type MenuItemResponse struct {
	ID           uint               `json:"id"`
	GroupID      uint               `json:"group_id"`
	ParentID     *uint              `json:"parent_id,omitempty"`
	Label        string             `json:"label"`
	Href         *string            `json:"href,omitempty"`
	Icon         string             `json:"icon"`
	Permission   *string            `json:"permission,omitempty"`
	DisplayOrder int                `json:"display_order"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Children     []MenuItemResponse `json:"children,omitempty"`
}

// SidebarGroup and SidebarItem for user-facing API (same as frontend expects)
type SidebarGroup struct {
	Group string        `json:"group"`
	Items []SidebarItem `json:"items"`
}

type SidebarItem struct {
	Label    string        `json:"label"`
	Href     string        `json:"href,omitempty"`
	Icon     string        `json:"icon"`
	Children []SidebarItem `json:"children,omitempty"`
}

type MenuListResponse struct {
	Items []models.MenuItem `json:"items"`
	BaseResponse
}
