package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// AdminWalletTxListFilters filters the admin wallet ledger list.
type AdminWalletTxListFilters struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Search   string `form:"search"`
	Type     string `form:"type"`
	Status   string `form:"status"`
	UserID   *uint  `form:"user_id"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

// AdminWalletTxListItem is a row in the admin wallet ledger table.
type AdminWalletTxListItem struct {
	ID              uint      `json:"id"`
	UserID          uint      `json:"user_id"`
	CustomerName    string    `json:"customer_name,omitempty"`
	CustomerEmail   string    `json:"customer_email,omitempty"`
	Amount          float64   `json:"amount"`
	Type            string    `json:"type"`
	ReferenceType   string    `json:"reference_type,omitempty"`
	ReferenceID     *uint     `json:"reference_id,omitempty"`
	Description     string    `json:"description,omitempty"`
	BalanceAfter    float64   `json:"balance_after"`
	Status          string    `json:"status"`
	StripeSessionID string    `json:"stripe_session_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// AdminWalletTxListData wraps the paginated admin wallet ledger response.
type AdminWalletTxListData struct {
	Transactions []AdminWalletTxListItem `json:"transactions"`
	Total        int64                   `json:"total"`
	Page         int                     `json:"page"`
	Limit        int                     `json:"limit"`
}

// AdminWalletTxDetailResponse powers the admin wallet transaction detail view.
type AdminWalletTxDetailResponse struct {
	ID              uint                   `json:"id"`
	UserID          uint                   `json:"user_id"`
	CustomerName    string                 `json:"customer_name,omitempty"`
	CustomerEmail   string                 `json:"customer_email,omitempty"`
	Amount          float64                `json:"amount"`
	Type            string                 `json:"type"`
	ReferenceType   string                 `json:"reference_type,omitempty"`
	ReferenceID     *uint                  `json:"reference_id,omitempty"`
	Description     string                 `json:"description,omitempty"`
	BalanceAfter    float64                `json:"balance_after"`
	Status          string                 `json:"status"`
	StripeSessionID string                 `json:"stripe_session_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// WalletTxTypeCount is a wallet transaction count bucketed by type.
type WalletTxTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// WalletTxSummaryResponse holds admin wallet ledger KPI counters.
type WalletTxSummaryResponse struct {
	TotalCount     int64               `json:"total_count"`
	CompletedCount int64               `json:"completed_count"`
	PendingCount   int64               `json:"pending_count"`
	FailedCount    int64               `json:"failed_count"`
	ByType         []WalletTxTypeCount `json:"by_type,omitempty"`
	NetVolume      float64             `json:"net_volume"`
}

// ToAdminWalletTxListItem maps a preloaded wallet transaction to a list row.
func ToAdminWalletTxListItem(tx *models.WalletTransaction) AdminWalletTxListItem {
	item := AdminWalletTxListItem{
		ID:              tx.ID,
		UserID:          tx.UserID,
		Amount:          tx.Amount,
		Type:            tx.Type,
		ReferenceType:   tx.ReferenceType,
		ReferenceID:     tx.ReferenceID,
		Description:     tx.Description,
		BalanceAfter:    tx.BalanceAfter,
		Status:          tx.Status,
		StripeSessionID: tx.StripeSessionID,
		CreatedAt:       tx.CreatedAt,
	}
	if tx.User.ID != 0 {
		item.CustomerName = strings.TrimSpace(tx.User.FirstName + " " + tx.User.LastName)
		item.CustomerEmail = tx.User.Email
	}
	return item
}

// ToAdminWalletTxDetail maps a fully preloaded wallet transaction to the admin detail response.
func ToAdminWalletTxDetail(tx *models.WalletTransaction) AdminWalletTxDetailResponse {
	resp := AdminWalletTxDetailResponse{
		ID:              tx.ID,
		UserID:          tx.UserID,
		Amount:          tx.Amount,
		Type:            tx.Type,
		ReferenceType:   tx.ReferenceType,
		ReferenceID:     tx.ReferenceID,
		Description:     tx.Description,
		BalanceAfter:    tx.BalanceAfter,
		Status:          tx.Status,
		StripeSessionID: tx.StripeSessionID,
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
	}
	if tx.User.ID != 0 {
		resp.CustomerName = strings.TrimSpace(tx.User.FirstName + " " + tx.User.LastName)
		resp.CustomerEmail = tx.User.Email
	}
	if len(tx.Metadata) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(tx.Metadata, &parsed); err == nil {
			resp.Metadata = parsed
		}
	}
	return resp
}
