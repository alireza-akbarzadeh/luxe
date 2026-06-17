package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type AdminServiceInterface interface {
	GetStats(ctx context.Context) (*dto.AdminStatsResponse, error)
}

type adminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) AdminServiceInterface {
	return &adminService{db: db}
}

func (s *adminService) GetStats(ctx context.Context) (*dto.AdminStatsResponse, error) {
	db := s.db.WithContext(ctx)

	var totalUsers int64
	if err := db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalOrders int64
	if err := db.Model(&models.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalProducts int64
	if err := db.Model(&models.Product{}).Where("status = ?", constants.ProductStatusActive).Count(&totalProducts).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalRevenue float64
	if err := db.Model(&models.Order{}).
		Where("status IN ?", []string{constants.OrderStatusPaid, constants.OrderStatusShipped, constants.OrderStatusDelivered}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var pendingOrders int64
	if err := db.Model(&models.Order{}).Where("status = ?", constants.OrderStatusPending).Count(&pendingOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalWalletBalance float64
	if err := db.Model(&models.Wallet{}).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&totalWalletBalance).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var lowStockProducts int64
	if err := db.Model(&models.Product{}).
		Where("stock <= low_stock_threshold AND status = ?", constants.ProductStatusActive).
		Count(&lowStockProducts).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.AdminStatsResponse{
		TotalUsers:        totalUsers,
		TotalOrders:       totalOrders,
		TotalActiveProducts: totalProducts,
		TotalRevenue:      totalRevenue,
		PendingOrders:     pendingOrders,
		TotalWalletBalance: totalWalletBalance,
		LowStockProducts:  lowStockProducts,
	}, nil
}
