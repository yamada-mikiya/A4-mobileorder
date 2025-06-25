package repositories

import "database/sql"

type AdminRepository interface {
	UpdateOrderStatus(orderID int, status string) error
}

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository (db *sql.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) UpdateOrderStatus (orderID int, status string) error {
	//TODO
	return nil
}