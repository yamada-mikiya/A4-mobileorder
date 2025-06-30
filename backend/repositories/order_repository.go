package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	GetProductList(shopID int) error
	CreateOrder(ctx context.Context, order *models.Order, products []models.OrderProduct) error
	ValidateAndGetProductsForShop(ctx context.Context, shopID int, productIDs []int) (map[int]models.Product, error)
	UpdateUserIDByGuestToken (ctx context.Context, guestToken string, userID int) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) GetProductList(shopID int) error {
	//TODO
	return nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order, products []models.OrderProduct) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("transaction rollback failed: %v, original error: %v", rbErr, err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	//orderにinsert
	orderQuery := `
			INSERT INTO orders (user_id, shop_id, order_date, total_amount, user_order_token, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING order_id, created_at, updated_at
	`
	err = tx.QueryRowContext(
		ctx,
		orderQuery,
		order.UserID,
		order.ShopID,
		time.Now(),
		order.TotalAmount,
		order.UserOrderToken,
		order.Status,
	).Scan(&order.OrderID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return err
	}

	//order_productにinsert
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO order_product (order_id, product_id, quantity, price_at_order) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, product := range products {
		if _, err = stmt.ExecContext(ctx, order.OrderID, product.ProductID, product.Quantity, product.PriceAtOrder); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderRepository) ValidateAndGetProductsForShop(ctx context.Context, shopID int, productIDs []int) (map[int]models.Product, error) {

	productMap := make(map[int]models.Product)

	if len(productIDs) == 0 {
		return productMap, nil
	}

	const baseQuery = `
		SELECT
			p.product_id,
			p.product_name,
			p.price,
			p.is_available
		FROM
			products p
		INNER JOIN
			shop_product sp ON p.product_id = sp.product_id
		WHERE
			sp.shop_id = ? AND p.product_id IN (?)
	`

	query, args, err := sqlx.In(baseQuery, shopID, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query for product validation: %w", err)
	}
	query = r.db.Rebind(query)

	var products []models.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select products for shop: %w", err)
	}

	if len(products) != len(productIDs) {
		return nil, errors.New("one or more products do not belong to the specified shop")
	}

	for _, p := range products {
		productMap[p.ProductID] = p
	}

	return productMap, nil

}

func (r *orderRepository) UpdateUserIDByGuestToken (ctx context.Context, guestToken string, userID int) error {
	query := "UPDATE orders SET user_id = $1 WHERE user_order_token = $2"
	result, err := r.db.ExecContext(ctx, query, userID, guestToken)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no matching guest order found for the provided token")
	}
	return nil
}