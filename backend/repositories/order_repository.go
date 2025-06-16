package repositories

import "database/sql"

type IOrderRepository interface {
	GetProductList(shopID int) error
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository (db *sql.DB) IOrderRepository {
	return &OrderRepository{db}
}

func (r *OrderRepository) GetProductList (shopID int) error {
	//TODO
	return nil
}