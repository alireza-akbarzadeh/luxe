package services

import (
	"io"

	importdata "github.com/alireza-akbarzadeh/luxe/internal/application/import"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
)

// ImportServiceInterface handles Excel-based bulk creation for admin import flows.
type ImportServiceInterface interface {
	ImportProductsFromExcel(r io.Reader, storeID uint) (*dto.ImportSummary, error)
	ImportCategoriesFromExcel(r io.Reader) (*dto.ImportSummary, error)
	ProductTemplate() ([]byte, error)
	CategoryTemplate() ([]byte, error)
}

type importService struct {
	inner *importdata.Service
}

func NewImportService(productSvc ProductServiceInterface, categorySvc CategoryServiceInterface) ImportServiceInterface {
	return &importService{
		inner: importdata.NewService(productSvc, categorySvc),
	}
}

func (s *importService) ImportProductsFromExcel(r io.Reader, storeID uint) (*dto.ImportSummary, error) {
	return s.inner.ImportProductsFromExcel(r, storeID)
}

func (s *importService) ImportCategoriesFromExcel(r io.Reader) (*dto.ImportSummary, error) {
	return s.inner.ImportCategoriesFromExcel(r)
}

func (s *importService) ProductTemplate() ([]byte, error) {
	return s.inner.ProductTemplate()
}

func (s *importService) CategoryTemplate() ([]byte, error) {
	return s.inner.CategoryTemplate()
}
