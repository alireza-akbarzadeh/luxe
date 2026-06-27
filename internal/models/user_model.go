// Package models defines the GORM data structures (User, RefreshToken) and their
// validation tags, corresponding to the PostgreSQL database schema.
package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	// Primary
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Account identity
	Email           string     `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	Phone     string `gorm:"index" json:"phone,omitempty" validate:"omitempty,e164"`
	FirstName string `gorm:"not null" json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `gorm:"not null" json:"last_name" validate:"required,min=1,max=100"`

	// Security & status
	PasswordHash    string `gorm:"not null" json:"-"`
	Role            string `gorm:"not null;default:'user';index" json:"role"`
	IsActive        bool   `gorm:"not null;default:true;index" json:"is_active"`
	WorkflowStateID *uint  `gorm:"index" json:"workflow_state_id,omitempty"`

	// Audit
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Legal acceptance (set at registration)
	TermsAcceptedAt   *time.Time `json:"terms_accepted_at,omitempty"`
	PrivacyAcceptedAt *time.Time `json:"privacy_accepted_at,omitempty"`
	TermsVersion      *string    `json:"terms_version,omitempty"`
	PrivacyVersion    *string    `json:"privacy_version,omitempty"`

	// Luxe Plus membership
	MembershipTier    string     `gorm:"not null;default:'free';index" json:"membership_tier"`
	PlusSubscribedAt  *time.Time `json:"plus_subscribed_at,omitempty"`
	PlusExpiresAt     *time.Time `json:"plus_expires_at,omitempty"`
}

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null;size:64"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}

type EmailVerificationToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null;size:64"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}
