package services

import (
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

type ItemServicer interface {
	GetItemList(shopID int) ([]models.ItemListResponse, error)
}

type itemService struct {
	r  repositories.ItemRepository
	db *sqlx.DB
}

func NewItemService(r repositories.ItemRepository, db *sqlx.DB) ItemServicer {
	return &itemService{r, db}
}

func (s *itemService) GetItemList(shopID int) ([]models.ItemListResponse, error) {
	itemList, err := s.r.GetItemList(s.db, shopID)
	if err != nil {
		return nil, err
	}
	return itemList, nil
}
