package dto

import (
	"encoding/json"
	"time"
)

// SetSettingRequest upsert a setting. The frontend sends any JSON value.
type SetSettingRequest struct {
	Value       json.RawMessage `json:"value" binding:"required" swaggertype:"object"`
	Description *string         `json:"description"`
}

// SettingResponse is what the API returns for any setting.
type SettingResponse struct {
	Key         string          `json:"key"`
	Value       json.RawMessage `json:"value" swaggertype:"object"`
	Description *string         `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
