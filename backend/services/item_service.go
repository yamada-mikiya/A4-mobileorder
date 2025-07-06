package services

import (
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
)

type ItemServicer interface {
	GetItemList(shopID int) ([]models.ItemListResponse, error)
}

type itemService struct {
	r repositories.ItemRepository
}

func NewItemService(r repositories.ItemRepository) ItemServicer {
	return &itemService{r}
}

func (s *itemService) GetItemList(shopID int) ([]models.ItemListResponse, error) {

	return nil, nil
}
