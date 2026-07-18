package userlike

import (
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

type Queries struct {
	repo *postgres.UserLikeRepository
}

// NewQueries creates user-like query use cases.
func NewQueries(repo *postgres.UserLikeRepository) *Queries {
	return &Queries{repo: repo}
}

// IsLikedByUser reports whether the user liked a product.
func (q *Queries) IsLikedByUser(userID, productID uint) (bool, error) {
	count, err := q.repo.CountLike(userID, productID)
	if err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

// GetUserLikedProductIDs returns product ids liked by the user.
func (q *Queries) GetUserLikedProductIDs(userID uint) ([]uint, error) {
	ids, err := q.repo.PluckLikedProductIDs(userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return ids, nil
}

// GetUserWishlist returns paginated wishlist products for a user.
func (q *Queries) GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error) {
	total, err := q.repo.CountWishlistProducts(userID)
	if err != nil {
		return nil, 0, err
	}

	orderByClause := wishlistOrderBy(sortBy)
	products, err := q.repo.FindWishlistProducts(userID, limit, offset, orderByClause)
	return products, total, err
}

func wishlistOrderBy(sortBy string) string {
	switch sortBy {
	case "price-asc":
		return "price ASC"
	case "price-desc":
		return "price DESC"
	case "name":
		return "name ASC"
	default:
		return "id DESC"
	}
}

// Commands orchestrates user-like write use cases.
type Commands struct {
	repo *postgres.UserLikeRepository
}

// NewCommands creates user-like command use cases.
func NewCommands(repo *postgres.UserLikeRepository) *Commands {
	return &Commands{repo: repo}
}

// Like adds a product to the user's wishlist.
func (c *Commands) Like(userID, productID uint) error {
	_, err := c.repo.FindProductByID(productID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("product not found")
		}
		return utils.ErrInternal(err)
	}

	like := &models.ProductLike{
		UserID:    userID,
		ProductID: productID,
	}

	err = c.repo.CreateLike(like)
	if err == nil {
		return nil
	}
	if isDuplicateKeyError(err) {
		return nil
	}
	return utils.ErrInternal(err)
}

// Unlike removes a product from the user's wishlist.
func (c *Commands) Unlike(userID, productID uint) error {
	if err := c.repo.DeleteLike(userID, productID); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
