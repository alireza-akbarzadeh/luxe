package models

import (
	"time"

	"gorm.io/gorm"
)

// SupportTicket is a customer support conversation (email, chat, or web).
type SupportTicket struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UserID        *uint      `gorm:"index" json:"user_id,omitempty"`
	CustomerName  string     `gorm:"not null;default:''" json:"customer_name,omitempty"`
	CustomerEmail string     `gorm:"not null;default:''" json:"customer_email,omitempty"`
	OrderID       *uint      `gorm:"index" json:"order_id,omitempty"`
	Subject       string     `gorm:"not null" json:"subject"`
	Status        string     `gorm:"not null;default:'open';index" json:"status"`
	Priority      string     `gorm:"not null;default:'normal'" json:"priority"`
	Channel       string     `gorm:"not null;default:'web';index" json:"channel"`
	AssigneeID    *uint      `gorm:"index" json:"assignee_id,omitempty"`
	AdminNotes    string     `json:"admin_notes,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`

	User     *User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Assignee *User                  `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Order    *Order                 `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Messages []SupportTicketMessage `gorm:"foreignKey:TicketID" json:"messages,omitempty"`
}

// SupportTicketMessage is one message in a support ticket thread.
type SupportTicketMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TicketID       uint   `gorm:"not null;index" json:"ticket_id"`
	AuthorID       *uint  `json:"author_id,omitempty"`
	AuthorRole     string `gorm:"not null;default:'customer'" json:"author_role"`
	Body           string `gorm:"not null" json:"body"`
	Channel        string `gorm:"not null;default:'web'" json:"channel"`
	IsInternal     bool   `gorm:"not null;default:false" json:"is_internal"`
	IsAISuggestion bool   `gorm:"not null;default:false" json:"is_ai_suggestion"`

	Author *User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}
