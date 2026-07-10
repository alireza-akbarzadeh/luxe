package dto

import "time"

type TeamResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description *string              `json:"description,omitempty"`
	MemberCount int64                `json:"member_count"`
	Members     []TeamMemberResponse `json:"members,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type TeamMemberResponse struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=120"`
	Slug        string `json:"slug" binding:"required,min=1,max=120"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=120"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type AddTeamMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=lead member"`
}
