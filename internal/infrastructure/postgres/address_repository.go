package postgres

import (
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AddressRepository implements address persistence with GORM.
type AddressRepository struct {
	db *gorm.DB
}

// NewAddressRepository creates a GORM-backed address repository.
func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

// CountByUser returns the number of addresses for a user.
func (r *AddressRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Address{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// Create inserts a new address.
func (r *AddressRepository) Create(address *models.Address) error {
	return r.db.Create(address).Error
}

// FindByIDAndUser loads an address owned by the user.
func (r *AddressRepository) FindByIDAndUser(id, userID uint) (*models.Address, error) {
	var addr models.Address
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&addr).Error; err != nil {
		return nil, err
	}
	return &addr, nil
}

// Save persists address changes.
func (r *AddressRepository) Save(addr *models.Address) error {
	return r.db.Save(addr).Error
}

// DeleteByIDAndUser removes an address owned by the user.
func (r *AddressRepository) DeleteByIDAndUser(id, userID uint) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Address{})
	return result.RowsAffected, result.Error
}

// ListByUser returns all addresses for a user ordered by default then created_at.
func (r *AddressRepository) ListByUser(userID uint) ([]models.Address, error) {
	var addresses []models.Address
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").Find(&addresses).Error
	return addresses, err
}

// UnsetDefault clears default flag for addresses matching type.
func (r *AddressRepository) UnsetDefault(userID uint, addressType string) error {
	return r.db.Model(&models.Address{}).
		Where("user_id = ? AND address_type IN (?, ?)", userID, addressType, "both").
		Where("address_type = ? OR address_type = 'both'", addressType).
		Update("is_default", false).Error
}

// SetDefault marks an address as default.
func (r *AddressRepository) SetDefault(id uint) error {
	return r.db.Model(&models.Address{}).Where("id = ?", id).Update("is_default", true).Error
}

// FindDefaultByUser loads the default address for a user and optional type.
func (r *AddressRepository) FindDefaultByUser(userID uint, addressType string) (*models.Address, error) {
	var addr models.Address
	query := r.db.Where("user_id = ? AND is_default = ?", userID, true)
	if addressType != "" {
		query = query.Where("address_type IN (?, ?)", addressType, "both")
	}
	if err := query.First(&addr).Error; err != nil {
		return nil, err
	}
	return &addr, nil
}
