package address

import (
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Commands orchestrates address write use cases.
type Commands struct {
	repo *postgres.AddressRepository
}

// NewCommands creates address command use cases.
func NewCommands(repo *postgres.AddressRepository) *Commands {
	return &Commands{repo: repo}
}

// Create inserts a new address for the user.
func (c *Commands) Create(userID uint, req dto.CreateAddressRequest) (*models.Address, error) {
	count, err := c.repo.CountByUser(userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if count == 0 {
		req.IsDefault = true
	}

	if req.IsDefault {
		if err := c.repo.UnsetDefault(userID, req.AddressType); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	address := &models.Address{
		UserID:        userID,
		AddressType:   req.AddressType,
		IsDefault:     req.IsDefault,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		AddressLine1:  req.AddressLine1,
		AddressLine2:  req.AddressLine2,
		City:          req.City,
		State:         req.State,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		Instructions:  req.Instructions,
	}

	if err := c.repo.Create(address); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return address, nil
}

// Update modifies an address owned by the user.
func (c *Commands) Update(id, userID uint, req dto.UpdateAddressRequest) (*models.Address, error) {
	addr, err := c.repo.FindByIDAndUser(id, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("address not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if req.AddressType != nil {
		addr.AddressType = *req.AddressType
	}
	if req.IsDefault != nil && *req.IsDefault {
		if err := c.repo.UnsetDefault(userID, addr.AddressType); err != nil {
			return nil, utils.ErrInternal(err)
		}
		addr.IsDefault = true
	} else if req.IsDefault != nil && !*req.IsDefault {
		addr.IsDefault = false
	}
	if req.RecipientName != nil {
		addr.RecipientName = *req.RecipientName
	}
	if req.Phone != nil {
		addr.Phone = *req.Phone
	}
	if req.AddressLine1 != nil {
		addr.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		addr.AddressLine2 = *req.AddressLine2
	}
	if req.City != nil {
		addr.City = *req.City
	}
	if req.State != nil {
		addr.State = *req.State
	}
	if req.PostalCode != nil {
		addr.PostalCode = *req.PostalCode
	}
	if req.Country != nil {
		addr.Country = *req.Country
	}
	if req.Instructions != nil {
		addr.Instructions = *req.Instructions
	}

	if err := c.repo.Save(addr); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return addr, nil
}

// Delete removes an address owned by the user.
func (c *Commands) Delete(id, userID uint) error {
	rows, err := c.repo.DeleteByIDAndUser(id, userID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("address not found")
	}
	return nil
}

// SetDefault marks an address as the default for its type.
func (c *Commands) SetDefault(id, userID uint) error {
	addr, err := c.repo.FindByIDAndUser(id, userID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("address not found")
		}
		return utils.ErrInternal(err)
	}
	if err := c.repo.UnsetDefault(userID, addr.AddressType); err != nil {
		return utils.ErrInternal(err)
	}
	return c.repo.SetDefault(id)
}
