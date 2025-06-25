package services

import "github.com/A4-dev-team/mobileorder.git/repositories"

type OrderServicer interface {
	GetProductListHandler(shopID int) error
}

type orderService struct {
	repo repositories.OrderRepository
}

func NewOrderService(r repositories.OrderRepository) OrderServicer {
	return &orderService{repo: r}
}

func (s *orderService) GetProductListHandler(shopID int) error {
	//TODO
	return nil
}
