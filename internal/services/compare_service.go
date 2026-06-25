package services

import (
	"context"

	appcompare "github.com/alireza-akbarzadeh/luxe/internal/application/compare"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"gorm.io/gorm"
)

type CompareServiceInterface interface {
	GetCompareList(userID uint) ([]uint, error)
	SyncCompareList(userID uint, productIDs []uint) error
	GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error)
}

type compareService struct {
	commands *appcompare.Commands
	queries  *appcompare.Queries
}

func NewCompareService(db *gorm.DB) CompareServiceInterface {
	repo := postgres.NewCompareRepository(db)
	return &compareService{
		commands: appcompare.NewCommands(repo),
		queries:  appcompare.NewQueries(repo),
	}
}

func (s *compareService) GetCompareList(userID uint) ([]uint, error) {
	return s.queries.GetCompareList(userID)
}

func (s *compareService) SyncCompareList(userID uint, productIDs []uint) error {
	return s.commands.SyncCompareList(userID, productIDs)
}

func (s *compareService) GetForCompare(ctx context.Context, productIDs []uint) ([]*dto.CompareProductResponse, error) {
	return s.queries.GetForCompare(ctx, productIDs)
}
