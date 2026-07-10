package models

import "time"

// Team groups staff users for access and workflow assignment.
type Team struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Members []TeamMember `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

// TeamMember links a user to a team with an optional role label (lead, member).
type TeamMember struct {
	TeamID    uint      `gorm:"primaryKey" json:"team_id"`
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	Role      string    `gorm:"not null;default:'member'" json:"role"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Team Team `gorm:"foreignKey:TeamID" json:"-"`
}
