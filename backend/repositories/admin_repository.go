package repositories

import "github.com/jmoiron/sqlx"

type AdminRepository interface {
	UpdateOrderStatus(orderID int, status string) error
}

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) UpdateOrderStatus(orderID int, status string) error {
	//TODO
	return nil
}