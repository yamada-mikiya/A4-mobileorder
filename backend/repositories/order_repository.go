package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order, products []models.OrderProduct) error
	UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db}
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

func (r *orderRepository) UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error {
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
