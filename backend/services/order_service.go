package services

import "github.com/A4-dev-team/mobileorder.git/repositories"

type IOrderServicer interface {
	GetProductListHandler(shopID int) error
}

type OrderService struct {
	repo repositories.IOrderRepository
}

func NewOrderService(r repositories.IOrderRepository) IOrderServicer {
	return &OrderService{repo: r}
}

func (s *OrderService) GetProductListHandler(shopID int) error {
	//TODO
	return nil
}
