package services

import (
	appaddress "github.com/alireza-akbarzadeh/luxe/internal/application/address"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type addressService struct {
	commands *appaddress.Commands
	queries  *appaddress.Queries
}

type AddressServiceInterface interface {
	Create(userID uint, req dto.CreateAddressRequest) (*models.Address, error)
	GetByID(id, userID uint) (*models.Address, error)
	Update(id, userID uint, req dto.UpdateAddressRequest) (*models.Address, error)
	Delete(id, userID uint) error
	List(userID uint) ([]models.Address, error)
	SetDefault(id, userID uint) error
	GetDefaultAddress(userID uint, addressType string) (*models.Address, error)
}

func NewAddressService(db *gorm.DB) AddressServiceInterface {
	repo := postgres.NewAddressRepository(db)
	return &addressService{
		commands: appaddress.NewCommands(repo),
		queries:  appaddress.NewQueries(repo),
	}
}

func (s *addressService) Create(userID uint, req dto.CreateAddressRequest) (*models.Address, error) {
	return s.commands.Create(userID, req)
}

func (s *addressService) GetByID(id, userID uint) (*models.Address, error) {
	return s.queries.GetByID(id, userID)
}

func (s *addressService) Update(id, userID uint, req dto.UpdateAddressRequest) (*models.Address, error) {
	return s.commands.Update(id, userID, req)
}

func (s *addressService) Delete(id, userID uint) error {
	return s.commands.Delete(id, userID)
}

func (s *addressService) List(userID uint) ([]models.Address, error) {
	return s.queries.List(userID)
}

func (s *addressService) SetDefault(id, userID uint) error {
	return s.commands.SetDefault(id, userID)
}

func (s *addressService) GetDefaultAddress(userID uint, addressType string) (*models.Address, error) {
	return s.queries.GetDefaultAddress(userID, addressType)
}
