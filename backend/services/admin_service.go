package services

import "github.com/A4-dev-team/mobileorder.git/repositories"

type AdminServicer interface {
	UpdateOrderStatus(orderID int, status string) error
}

type adminService struct {
	repo repositories.AdminRepository
}

func NewAdminService(r repositories.AdminRepository) AdminServicer {
	return &adminService{repo: r}
}

func (s *adminService) UpdateOrderStatus(orderID int, status string) error {
	//TODO
	return nil
}
