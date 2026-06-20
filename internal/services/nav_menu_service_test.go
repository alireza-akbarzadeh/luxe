package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNavMenuService_Reorder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)

	svc := NewNavMenuService(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "nav_menus" SET "order"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(2, sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "nav_menus" SET "order"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(1, sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = svc.Reorder(context.Background(), &dto.ReorderNavMenusRequest{
		Items: []dto.ReorderNavMenuItem{
			{ID: 1, Order: 2},
			{ID: 2, Order: 1},
		},
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
