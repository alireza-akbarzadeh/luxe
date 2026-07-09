package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ReturnRepository persists return requests with GORM.
type ReturnRepository struct {
	db *gorm.DB
}

// NewReturnRepository creates a GORM-backed return repository.
func NewReturnRepository(db *gorm.DB) *ReturnRepository {
	return &ReturnRepository{db: db}
}

// FindUserOrder loads an order owned by a user.
func (r *ReturnRepository) FindUserOrder(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// CountOpenForOrder counts non-terminal returns for an order/user pair.
func (r *ReturnRepository) CountOpenForOrder(ctx context.Context, orderID, userID uint, excludedStatuses []string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Return{}).
		Where("order_id = ? AND user_id = ? AND status NOT IN ?", orderID, userID, excludedStatuses).
		Count(&count).Error
	return count, err
}

// Create inserts a return record.
func (r *ReturnRepository) Create(ctx context.Context, ret *models.Return) error {
	return r.db.WithContext(ctx).Create(ret).Error
}

// FindByID loads a return with optional user scope.
func (r *ReturnRepository) FindByID(ctx context.Context, returnID uint, userID uint, isAdmin bool) (*models.Return, error) {
	q := r.db.WithContext(ctx).
		Preload("Order").
		Preload("User").
		Preload("WorkflowState").
		Where("id = ?", returnID)
	if !isAdmin {
		q = q.Where("user_id = ?", userID)
	}
	var ret models.Return
	if err := q.First(&ret).Error; err != nil {
		return nil, err
	}
	return &ret, nil
}

// CountForUser counts returns for a user.
func (r *ReturnRepository) CountForUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Return{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

// ListForUser returns paginated returns for a user.
func (r *ReturnRepository) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, error) {
	var returns []models.Return
	err := r.db.WithContext(ctx).Model(&models.Return{}).Where("user_id = ?", userID).
		Preload("Order").Preload("User").Preload("WorkflowState").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&returns).Error
	return returns, err
}

func (r *ReturnRepository) applyAdminFilters(q *gorm.DB, filters dto.AdminReturnListFilters) *gorm.DB {
	if filters.Status != "" {
		q = q.Where("returns.status = ?", filters.Status)
	}
	if filters.ReturnType != "" {
		q = q.Where("returns.return_type = ?", filters.ReturnType)
	}
	if filters.WorkflowState != "" {
		q = q.Joins("JOIN workflow_states ws ON ws.id = returns.workflow_state_id").
			Where("ws.code = ?", filters.WorkflowState)
	}
	if filters.UserID != nil {
		q = q.Where("returns.user_id = ?", *filters.UserID)
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		q = q.Joins("LEFT JOIN orders ON orders.id = returns.order_id AND orders.deleted_at IS NULL").
			Joins("LEFT JOIN users ON users.id = returns.user_id AND users.deleted_at IS NULL").
			Where(
				"returns.reason ILIKE ? OR orders.order_number ILIKE ? OR users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?",
				term, term, term, term, term,
			)
	}
	return q
}

// CountAdmin counts returns matching admin filters.
func (r *ReturnRepository) CountAdmin(ctx context.Context, filters dto.AdminReturnListFilters) (int64, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Return{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdmin returns paginated returns for admin.
func (r *ReturnRepository) ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters, limit, offset int) ([]models.Return, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Return{}), filters)
	var returns []models.Return
	err := q.Preload("Order").Preload("User").Preload("WorkflowState").
		Order("returns.created_at DESC").Limit(limit).Offset(offset).
		Find(&returns).Error
	return returns, err
}

// Update persists return changes.
func (r *ReturnRepository) Update(ctx context.Context, ret *models.Return) error {
	return r.db.WithContext(ctx).Save(ret).Error
}

var adminReturnOpenStatuses = []string{"requested", "approved", "item_received", "refund_processing", "exchange_processing"}

// AdminStats aggregates return metrics for the admin dashboard.
func (r *ReturnRepository) AdminStats(ctx context.Context) (dto.AdminReturnStats, error) {
	stats := dto.AdminReturnStats{
		ByStatus: make(map[string]int64),
		ByType:   make(map[string]int64),
	}

	if err := r.db.WithContext(ctx).Model(&models.Return{}).Count(&stats.Total).Error; err != nil {
		return stats, err
	}

	if err := r.db.WithContext(ctx).Model(&models.Return{}).
		Where("status IN ?", adminReturnOpenStatuses).
		Count(&stats.Open).Error; err != nil {
		return stats, err
	}

	if err := r.db.WithContext(ctx).Model(&models.Return{}).
		Where("status = ?", "refunded").
		Select("COALESCE(SUM(refund_amount), 0)").
		Scan(&stats.RefundTotal).Error; err != nil {
		return stats, err
	}

	type statusRow struct {
		Status string
		Count  int64
	}
	var statusRows []statusRow
	if err := r.db.WithContext(ctx).Model(&models.Return{}).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&statusRows).Error; err != nil {
		return stats, err
	}
	for _, row := range statusRows {
		stats.ByStatus[row.Status] = row.Count
	}

	type typeRow struct {
		ReturnType string
		Count      int64
	}
	var typeRows []typeRow
	if err := r.db.WithContext(ctx).Model(&models.Return{}).
		Select("return_type, COUNT(*) AS count").
		Group("return_type").
		Scan(&typeRows).Error; err != nil {
		return stats, err
	}
	for _, row := range typeRows {
		stats.ByType[row.ReturnType] = row.Count
	}

	type dayRow struct {
		Day   string
		Count int64
	}
	var dayRows []dayRow
	if err := r.db.WithContext(ctx).Model(&models.Return{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS day, COUNT(*) AS count").
		Where("created_at >= NOW() - INTERVAL '7 days'").
		Group("day").
		Order("day ASC").
		Scan(&dayRows).Error; err != nil {
		return stats, err
	}

	stats.Last7Days = make([]dto.ReturnDailyCount, 0, len(dayRows))
	for _, row := range dayRows {
		stats.Last7Days = append(stats.Last7Days, dto.ReturnDailyCount{
			Date:  row.Day,
			Count: row.Count,
		})
	}

	return stats, nil
}

// ProductReturnStats aggregates order and return signals for a catalog product.
func (r *ReturnRepository) ProductReturnStats(ctx context.Context, productID uint) (orderCount int64, returnCount int64, reasons []string, err error) {
	if productID == 0 {
		return 0, 0, nil, nil
	}

	err = r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Where("product_id = ?", productID).
		Distinct("order_id").
		Count(&orderCount).Error
	if err != nil {
		return 0, 0, nil, err
	}

	err = r.db.WithContext(ctx).Model(&models.Return{}).
		Joins("JOIN order_items ON order_items.order_id = returns.order_id AND order_items.deleted_at IS NULL").
		Where("order_items.product_id = ?", productID).
		Distinct("returns.id").
		Count(&returnCount).Error
	if err != nil {
		return 0, 0, nil, err
	}

	type reasonCount struct {
		Reason string
		Count  int64
	}
	var rows []reasonCount
	err = r.db.WithContext(ctx).Model(&models.Return{}).
		Select("returns.reason AS reason, COUNT(DISTINCT returns.id) AS count").
		Joins("JOIN order_items ON order_items.order_id = returns.order_id AND order_items.deleted_at IS NULL").
		Where("order_items.product_id = ? AND returns.reason <> ''", productID).
		Group("returns.reason").
		Order("count DESC").
		Limit(5).
		Scan(&rows).Error
	if err != nil {
		return orderCount, returnCount, nil, err
	}

	reasons = make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Reason != "" {
			reasons = append(reasons, row.Reason)
		}
	}
	return orderCount, returnCount, reasons, nil
}
