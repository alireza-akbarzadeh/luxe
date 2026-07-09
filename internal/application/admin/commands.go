package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates admin write use cases.
type Commands struct {
	repo    *postgres.AdminRepository
	navRepo *postgres.AdminNavRepository
}

// NewCommands creates admin command use cases.
func NewCommands(repo *postgres.AdminRepository, navRepo *postgres.AdminNavRepository) *Commands {
	return &Commands{repo: repo, navRepo: navRepo}
}

// UpdateUserRole sets a user's role after validation by caller.
func (c *Commands) UpdateUserRole(ctx context.Context, userID uint, role string) error {
	rows, err := c.repo.UpdateUserRole(ctx, userID, role)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("user not found")
	}
	return nil
}

// ToggleUserActive sets a user's active flag.
func (c *Commands) ToggleUserActive(ctx context.Context, userID uint, active bool) error {
	rows, err := c.repo.UpdateUserActive(ctx, userID, active)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("user not found")
	}
	return nil
}

// ExportOrdersCSV builds a CSV export for orders.
func (c *Commands) ExportOrdersCSV(ctx context.Context, filters dto.AdminOrderExportFilters) ([]byte, error) {
	orders, err := c.repo.ListOrdersForExport(ctx, filters)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "order_number", "status", "total_amount", "currency", "user_email", "created_at"})
	for _, o := range orders {
		email := ""
		if o.User.Email != "" {
			email = o.User.Email
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", o.ID),
			o.OrderNumber,
			o.Status,
			fmt.Sprintf("%.2f", o.TotalAmount),
			o.Currency,
			email,
			o.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return buf.Bytes(), nil
}
