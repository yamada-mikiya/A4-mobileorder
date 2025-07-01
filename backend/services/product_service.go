package services

import (
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
)

type ProductServicer interface {
	GetProductList(shopID int) ([]models.ProductListResponse, error)
}

type productService struct {
	r repositories.ProductRepository
}

func NewProductService(r repositories.ProductRepository) ProductServicer {
	return &productService{r}
}

func (s *productService) GetProductList(shopID int) ([]models.ProductListResponse, error) {

	return nil, nil
}
