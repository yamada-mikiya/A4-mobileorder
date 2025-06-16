package services

import "github.com/A4-dev-team/mobileorder.git/repositories"

type IAdminServicer interface {
	UpdateOrderStatus(orderID int, status string) error
}

type AdminService struct {
	repo repositories.IAdminRepository
}

func NewAdminService(r repositories.IAdminRepository) IAdminServicer {
	return &AdminService{repo: r}
}

func (s *AdminService) UpdateOrderStatus(orderID int, status string) error {
	//TODO
	return nil
}
