package repositories

import "database/sql"

type IAdminRepository interface {
	UpdateOrderStatus(orderID int, status string) error
}

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository (db *sql.DB) IAdminRepository {
	return &AdminRepository{db}
}

func (r *AdminRepository) UpdateOrderStatus (orderID int, status string) error {
	//TODO
	return nil
}