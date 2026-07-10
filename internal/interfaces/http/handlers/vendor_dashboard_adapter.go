package handlers

import (
	"context"
	"fmt"
	"time"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	appcatalog "github.com/alireza-akbarzadeh/luxe/internal/application/catalog"
	orderfacade "github.com/alireza-akbarzadeh/luxe/internal/application/order/facade"
	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

type vendorDashboardAdapter struct {
	stores      *appstore.Queries
	catalog     *appcatalog.Service
	orders      *orderfacade.Service
	productRepo *postgres.ProductRepository
}

func (a vendorDashboardAdapter) LoadSnapshot(ctx context.Context, storeID uint) (*appai.VendorDashboardSnapshot, error) {
	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	orderStats, err := a.orders.GetVendorStoreOrderStats(ctx, storeID)
	if err != nil {
		return nil, err
	}

	productStats, err := a.catalog.GetVendorStoreProductStats(ctx, storeID)
	if err != nil {
		return nil, err
	}

	recentOrders, _, err := a.orders.ListVendorStoreOrders(ctx, storeID, apporder.AdminOrderFilters{}, 5, 0)
	if err != nil {
		return nil, err
	}

	lowStockProducts, err := a.productRepo.ListLowStockByStore(ctx, storeID, 5)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	snapshot := &appai.VendorDashboardSnapshot{
		StoreName:        store.Name,
		OrderTotal:       orderStats.Total,
		OrdersByStatus:   orderStats.ByStatus,
		ProductTotal:     productStats.Total,
		ProductsByStatus: productStats.ByStatus,
		LowStockCount:    productStats.LowStock,
		RecentOrders:     make([]appai.VendorDashboardOrderLine, 0, len(recentOrders)),
		LowStockItems:    make([]appai.VendorDashboardStockLine, 0, len(lowStockProducts)),
	}

	for _, order := range recentOrders {
		row := dto.ToVendorOrderListItem(order, storeID)
		snapshot.RecentOrders = append(snapshot.RecentOrders, appai.VendorDashboardOrderLine{
			OrderNumber: row.OrderNumber,
			Status:      row.Status,
			Subtotal:    row.StoreSubtotal,
			ItemCount:   row.StoreItemsCount,
		})
	}

	for _, product := range lowStockProducts {
		snapshot.LowStockItems = append(snapshot.LowStockItems, appai.VendorDashboardStockLine{
			Name:  product.Name,
			Stock: product.Stock,
		})
	}

	return snapshot, nil
}

func (a vendorDashboardAdapter) LoadSalesSnapshot(ctx context.Context, storeID uint, days int) (*appai.VendorSalesSnapshot, error) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	now := time.Now().UTC()
	periodEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	periodStart := periodEnd.AddDate(0, 0, -days)
	priorEnd := periodStart
	priorStart := priorEnd.AddDate(0, 0, -days)

	current, err := a.orders.GetVendorStoreSalesSummary(ctx, storeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	prior, err := a.orders.GetVendorStoreSalesSummary(ctx, storeID, priorStart, priorEnd)
	if err != nil {
		return nil, err
	}
	daily, err := a.orders.ListVendorDailySales(ctx, storeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	topProducts, err := a.orders.ListVendorTopProducts(ctx, storeID, periodStart, periodEnd, 5)
	if err != nil {
		return nil, err
	}

	return &appai.VendorSalesSnapshot{
		StoreName:   store.Name,
		PeriodDays:  days,
		Current:     current,
		Prior:       prior,
		Daily:       daily,
		TopProducts: topProducts,
	}, nil
}

func (a vendorDashboardAdapter) periodRange(days int) (time.Time, time.Time) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}
	now := time.Now().UTC()
	periodEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	periodStart := periodEnd.AddDate(0, 0, -days)
	return periodStart, periodEnd
}

func (a vendorDashboardAdapter) mapDemandRows(rows []postgres.ProductDemandRow) []appai.VendorProductDemand {
	products := make([]appai.VendorProductDemand, len(rows))
	for i, row := range rows {
		products[i] = appai.VendorProductDemand{
			ProductID:         row.ProductID,
			Name:              row.Name,
			Price:             row.Price,
			CompareAtPrice:    row.CompareAtPrice,
			Cost:              row.Cost,
			Stock:             row.Stock,
			LowStockThreshold: row.LowStockThreshold,
			UnitsSold:         row.UnitsSold,
			Revenue:           row.Revenue,
		}
	}
	return products
}

func (a vendorDashboardAdapter) LoadInventorySnapshot(ctx context.Context, storeID uint, days int) (*appai.VendorInventorySnapshot, error) {
	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	productStats, err := a.catalog.GetVendorStoreProductStats(ctx, storeID)
	if err != nil {
		return nil, err
	}

	from, to := a.periodRange(days)
	rows, err := a.productRepo.ListDemandSignalsByStore(ctx, storeID, from, to, true, 50)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &appai.VendorInventorySnapshot{
		StoreName:     store.Name,
		PeriodDays:    days,
		LowStockCount: productStats.LowStock,
		Products:      a.mapDemandRows(rows),
	}, nil
}

func (a vendorDashboardAdapter) LoadPricingSnapshot(ctx context.Context, storeID uint, days int) (*appai.VendorPricingSnapshot, error) {
	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	from, to := a.periodRange(days)
	rows, err := a.productRepo.ListDemandSignalsByStore(ctx, storeID, from, to, false, 50)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &appai.VendorPricingSnapshot{
		StoreName:  store.Name,
		PeriodDays: days,
		Products:   a.mapDemandRows(rows),
	}, nil
}

func (a vendorDashboardAdapter) LoadCustomerSnapshot(ctx context.Context, storeID uint, days int) (*appai.VendorCustomerSnapshot, error) {
	if days <= 0 {
		days = 365
	}
	if days > 730 {
		days = 730
	}

	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	from, to := a.periodRange(days)
	customers, err := a.orders.ListVendorStoreCustomers(ctx, storeID, from, to, 100)
	if err != nil {
		return nil, err
	}

	return &appai.VendorCustomerSnapshot{
		StoreName:  store.Name,
		PeriodDays: days,
		Customers:  customers,
	}, nil
}

func (a vendorDashboardAdapter) LoadAdminPerformance(ctx context.Context, storeID uint, days int) (*dto.AdminVendorPerformanceResponse, error) {
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	store, err := a.stores.GetByID(storeID)
	if err != nil {
		return nil, utils.ErrNotFound("store not found")
	}

	periodLabel := fmt.Sprintf("%dd", days)
	switch days {
	case 7:
		periodLabel = "7d"
	case 90:
		periodLabel = "90d"
	default:
		periodLabel = "30d"
	}

	now := time.Now().UTC()
	periodEnd := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	periodStart := periodEnd.AddDate(0, 0, -days)
	priorEnd := periodStart
	priorStart := priorEnd.AddDate(0, 0, -days)

	current, err := a.orders.GetVendorStoreSalesSummary(ctx, storeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	prior, err := a.orders.GetVendorStoreSalesSummary(ctx, storeID, priorStart, priorEnd)
	if err != nil {
		return nil, err
	}
	daily, err := a.orders.ListVendorDailySales(ctx, storeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	topProducts, err := a.orders.ListVendorTopProducts(ctx, storeID, periodStart, periodEnd, 8)
	if err != nil {
		return nil, err
	}
	orderStats, err := a.orders.GetVendorStoreOrderStats(ctx, storeID)
	if err != nil {
		return nil, err
	}
	productStats, err := a.catalog.GetVendorStoreProductStats(ctx, storeID)
	if err != nil {
		return nil, err
	}

	topRows := make([]dto.AdminVendorTopProduct, len(topProducts))
	for i, row := range topProducts {
		topRows[i] = dto.AdminVendorTopProduct{
			ProductID: row.ProductID,
			Name:      row.Name,
			Revenue:   row.Revenue,
			Units:     row.Units,
		}
	}

	dailyRows := make([]dto.AdminVendorDailySales, len(daily))
	for i, row := range daily {
		dailyRows[i] = dto.AdminVendorDailySales{
			Date:    row.Date,
			Revenue: row.Revenue,
			Orders:  row.Orders,
		}
	}

	return &dto.AdminVendorPerformanceResponse{
		Period:      periodLabel,
		GeneratedAt: now,
		Store:       dto.ToAdminStoreResponse(ctx, store),
		CurrentSales: dto.AdminVendorSalesSummary{
			Revenue:       current.Revenue,
			OrderCount:    current.OrderCount,
			UnitsSold:     current.UnitsSold,
			AvgOrderValue: current.AvgOrderValue,
		},
		PreviousSales: dto.AdminVendorSalesSummary{
			Revenue:       prior.Revenue,
			OrderCount:    prior.OrderCount,
			UnitsSold:     prior.UnitsSold,
			AvgOrderValue: prior.AvgOrderValue,
		},
		OrderTotal:       orderStats.Total,
		OrdersByStatus:   orderStats.ByStatus,
		ProductTotal:     productStats.Total,
		ProductsByStatus: productStats.ByStatus,
		LowStockCount:    productStats.LowStock,
		TopProducts:      topRows,
		DailySales:       dailyRows,
	}, nil
}
