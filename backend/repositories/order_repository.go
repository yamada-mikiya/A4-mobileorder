package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order, products []models.OrderProduct) error
	UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error
	FindUserOrdersWithDetails(ctx context.Context, userID int, statuses []models.OrderStatus) ([]OrderWithDetailsDB, error)
	FindProductsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.OrderItem, error)
	FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error)
	CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error)
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
			INSERT INTO orders (user_id, shop_id, order_date, total_amount, guest_order_token, status)
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
		order.GuestOrderToken,
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
	query := "UPDATE orders SET user_id = $1 WHERE guest_order_token = $2"
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

// order情報をとってくる。
type OrderWithDetailsDB struct {
	OrderID      int                `db:"order_id"`
	ShopName     string             `db:"shop_name"`
	OrderDate    time.Time          `db:"order_date"`
	TotalAmount  float64            `db:"total_amount"`
	Status       models.OrderStatus `db:"status"`
	WaitingCount int                `db:"waiting_count"`
}

func (r *orderRepository) FindUserOrdersWithDetails(ctx context.Context, userID int, statuses []models.OrderStatus) ([]OrderWithDetailsDB, error) {
	if len(statuses) == 0 {
		return []OrderWithDetailsDB{}, nil
	}
	statusInts := make([]interface{}, len(statuses))
	for i, s := range statuses {
		statusInts[i] = int(s)
	}
	query, args, err := sqlx.In(`
		SELECT
			o.order_id,
			s.name AS shop_name,
			o.order_date,
			o.total_amount,
			o.status,
			CASE
				WHEN o.status = ? THEN
					(SELECT COUNT(*)
					 FROM orders sub
					 WHERE sub.shop_id = o.shop_id AND status = ? AND sub.order_date < o.order_date)
				ELSE 0
			END AS waiting_count
		FROM
			orders o
		INNER JOIN
			shops s ON o.shop_id = s.shop_id
		WHERE
			o.user_id = ? AND o.status IN (?)
		ORDER BY
			o.order_date DESC;
	`, models.Cooking, models.Cooking, userID, statusInts)

	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	query = r.db.Rebind(query)

	var orders []OrderWithDetailsDB
	if err := r.db.SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select user orders: %w", err)
	}

	return orders, nil
}

// 注文の商品が何なのかとってくる
func (r *orderRepository) FindProductsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.OrderItem, error) {
	if len(orderIDs) == 0 {
		return make(map[int][]models.OrderItem), nil
	}

	query, args, err := sqlx.In(`
		SELECT op.order_id, p.product_name, op.quantity
		FROM order_product op
		INNER JOIN products p ON op.product_id = p.Product_id
		WHERE op.order_id IN (?)
	`, orderIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	itemsMap := make(map[int][]models.OrderItem)
	for rows.Next() {
		var orderID int
		var item models.OrderItem
		if err := rows.Scan(&orderID, &item.ProductName, &item.Quantity); err != nil {
			return nil, err
		}
		itemsMap[orderID] = append(itemsMap[orderID], item)
	}

	return itemsMap, nil
}

func (r *orderRepository) FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	var order models.Order
	query := "SELECT * FROM orders WHERE order_id = $1 AND user_id = $2"
	if err := r.db.GetContext(ctx, &order, query, orderID, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found or you do not have permission")
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE shop_id = $1 AND status = $2 AND order_date < $3`
	err := r.db.GetContext(ctx, &count, query, shopID, models.Cooking, orderDate)
	return count, err
}

