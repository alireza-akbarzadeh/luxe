package models

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
)

// WebhookEvent stores every inbound webhook for idempotency checking and debugging.
type WebhookEvent struct {
	ID          uint       `gorm:"primaryKey"                          json:"id"`
	EventID     string     `gorm:"uniqueIndex;not null"                json:"event_id"`
	EventType   string     `gorm:"not null;index"                      json:"event_type"`
	Source      string     `gorm:"not null;default:'stripe'"           json:"source"`
	Status      string     `gorm:"not null;default:'received';index"   json:"status"`
	Payload     []byte     `gorm:"type:jsonb;not null;default:'{}'"    json:"-"`
	ErrorMsg    string     `gorm:"column:error_msg"                    json:"error_msg,omitempty"`
	CreatedAt   time.Time  `                                           json:"created_at"`
	ProcessedAt *time.Time `gorm:"index"                               json:"processed_at,omitempty"`
}

// IsAlreadyProcessed returns true when the event reached a terminal state.
func (e *WebhookEvent) IsAlreadyProcessed() bool {
	return e.Status == constants.WebhookStatusProcessed || e.Status == constants.WebhookStatusFailed
}
