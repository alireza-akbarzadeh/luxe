package address

import (
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates address read use cases.
type Queries struct {
	repo *postgres.AddressRepository
}

// NewQueries creates address query use cases.
func NewQueries(repo *postgres.AddressRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID returns an address owned by the user.
func (q *Queries) GetByID(id, userID uint) (*models.Address, error) {
	addr, err := q.repo.FindByIDAndUser(id, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("address not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return addr, nil
}

// List returns all addresses for a user.
func (q *Queries) List(userID uint) ([]models.Address, error) {
	addresses, err := q.repo.ListByUser(userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return addresses, nil
}

// GetDefaultAddress returns the default address or nil when none exists.
func (q *Queries) GetDefaultAddress(userID uint, addressType string) (*models.Address, error) {
	addr, err := q.repo.FindDefaultByUser(userID, addressType)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return addr, nil
}
