package repositories

import "database/sql"

type OrderRepository interface {
	GetProductList(shopID int) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository (db *sql.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) GetProductList (shopID int) error {
	//TODO
	return nil
}