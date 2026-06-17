package services

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminService_GetStats(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewAdminService(db, nil)

	// Each Count/Scan call translates to a SELECT; match them in order.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(30))

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(total_amount\), 0\)`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(9999.50))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(balance\), 0\)`).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(500.00))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	stats, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(42), stats.TotalUsers)
	assert.Equal(t, int64(100), stats.TotalOrders)
	assert.Equal(t, int64(30), stats.TotalActiveProducts)
	assert.InDelta(t, 9999.50, stats.TotalRevenue, 0.01)
	assert.Equal(t, int64(5), stats.PendingOrders)
	assert.InDelta(t, 500.00, stats.TotalWalletBalance, 0.01)
	assert.Equal(t, int64(3), stats.LowStockProducts)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminService_UpdateUserRole_InvalidRole(t *testing.T) {
	db, _ := setupTestDB(t)
	svc := NewAdminService(db, nil)

	err := svc.UpdateUserRole(context.Background(), 1, "superuser")
	require.Error(t, err)
}

func TestAdminService_UpdateUserRole_ValidRoles(t *testing.T) {
	for _, role := range []string{constants.RoleAdmin, constants.RoleUser} {
		t.Run(role, func(t *testing.T) {
			db, mock := setupTestDB(t)
			svc := NewAdminService(db, nil)

			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE "users"`).
				WithArgs(role, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			err := svc.UpdateUserRole(context.Background(), 1, role)
			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAdminService_ListUsers_EmptyFilters(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewAdminService(db, nil)

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "role", "is_active", "created_at", "updated_at"}).
			AddRow(1, "a@test.com", "Alice", "A", "user", true, now, now).
			AddRow(2, "b@test.com", "Bob", "B", "admin", true, now, now))

	filters := dto.AdminUserFilters{Limit: 20, Offset: 0}
	users, total, err := svc.ListUsers(context.Background(), filters)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.Equal(t, "a@test.com", users[0].Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminService_ToggleUserActive(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewAdminService(db, nil)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WithArgs(false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := svc.ToggleUserActive(context.Background(), 1, false)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
