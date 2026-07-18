package dto

import "time"

type RoleResponse struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description,omitempty"`
	IsSystem        bool      `json:"is_system"`
	UserCount       int64     `json:"user_count"`
	PermissionCount int64     `json:"permission_count"`
	PermissionKeys  []string  `json:"permission_keys,omitempty"`
	PermissionIDs   []uint    `json:"permission_ids,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PermissionResponse struct {
	ID          uint    `json:"id"`
	Key         string  `json:"key"`
	Module      string  `json:"module"`
	Description *string `json:"description,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=80"`
	Slug        string `json:"slug" binding:"required,min=1,max=80"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=80"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type SetRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids"`
}
