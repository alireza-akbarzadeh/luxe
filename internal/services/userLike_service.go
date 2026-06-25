package services

import (
	appuserlike "github.com/alireza-akbarzadeh/luxe/internal/application/userlike"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type UsertLikeServiceInterface interface {
	Like(userID, productID uint) error
	Unlike(userID, productID uint) error
	IsLikedByUser(userID, productID uint) (bool, error)
	GetUserLikedProductIDs(userID uint) ([]uint, error)
	GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error)
}

type productLikeService struct {
	commands *appuserlike.Commands
	queries  *appuserlike.Queries
}

func NewUserLikeService(db *gorm.DB) UsertLikeServiceInterface {
	repo := postgres.NewUserLikeRepository(db)
	return &productLikeService{
		commands: appuserlike.NewCommands(repo),
		queries:  appuserlike.NewQueries(repo),
	}
}

func (s *productLikeService) Like(userID, productID uint) error {
	return s.commands.Like(userID, productID)
}

func (s *productLikeService) Unlike(userID, productID uint) error {
	return s.commands.Unlike(userID, productID)
}

func (s *productLikeService) IsLikedByUser(userID, productID uint) (bool, error) {
	return s.queries.IsLikedByUser(userID, productID)
}

func (s *productLikeService) GetUserLikedProductIDs(userID uint) ([]uint, error) {
	return s.queries.GetUserLikedProductIDs(userID)
}

func (s *productLikeService) GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error) {
	return s.queries.GetUserWishlist(userID, limit, offset, sortBy)
}
