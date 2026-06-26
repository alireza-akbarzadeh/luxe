package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type productStatusCountRow struct {
	Status string
	Count  int64
}

// CountByStoreStatus returns product totals, status breakdown, and low-stock count for a store.
func (r *ProductRepository) CountByStoreStatus(ctx context.Context, storeID uint) (dto.VendorProductStats, error) {
	stats := dto.VendorProductStats{ByStatus: map[string]int64{}}

	base := r.db.WithContext(ctx).Model(&models.Product{}).Where("store_id = ?", storeID)

	if err := base.Count(&stats.Total).Error; err != nil {
		return stats, err
	}

	var rows []productStatusCountRow
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("status, COUNT(*) AS count").
		Where("store_id = ?", storeID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return stats, err
	}

	for _, row := range rows {
		stats.ByStatus[row.Status] = row.Count
	}

	err = r.db.WithContext(ctx).Model(&models.Product{}).
		Where("store_id = ? AND status = ? AND track_inventory = ? AND stock <= low_stock_threshold AND stock > 0",
			storeID, constants.ProductStatusActive, true).
		Count(&stats.LowStock).Error
	if err != nil {
		return stats, err
	}

	return stats, nil
}
