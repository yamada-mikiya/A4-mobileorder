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
	CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) error
	UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error
	FindActiveUserOrders(ctx context.Context, userID int) ([]OrderWithDetailsDB, error)
	FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error)
	FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error)
	CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error)
	FindActiveShopOrders(ctx context.Context, shopID int) ([]AdminOrderDBResult, error)
	FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error
	DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) (err error) {
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

	//order_itemにinsert
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO order_item (order_id, item_id, quantity, price_at_order) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.ExecContext(ctx, order.OrderID, item.ItemID, item.Quantity, item.PriceAtOrder); err != nil {
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

// ユーザーの注文情報をとってくる。
type OrderWithDetailsDB struct {
	OrderID      int                `db:"order_id"`
	ShopName     string             `db:"shop_name"`
	Location     string             `db:"location"`
	OrderDate    time.Time          `db:"order_date"`
	TotalAmount  float64            `db:"total_amount"`
	Status       models.OrderStatus `db:"status"`
	WaitingCount int                `db:"waiting_count"`
}

func (r *orderRepository) FindActiveUserOrders(ctx context.Context, userID int) ([]OrderWithDetailsDB, error) {
	query := `
		SELECT
			o.order_id,
			s.name AS shop_name,
			s.location,
			o.order_date,
			o.total_amount,
			o.status,
			CASE
				WHEN o.status = $1 THEN
					(SELECT COUNT(*)
					 FROM orders sub
					 WHERE sub.shop_id = o.shop_id AND sub.status = $1 AND sub.order_date < o.order_date)
				ELSE 0
			END AS waiting_count
		FROM
			orders o
		INNER JOIN
			shops s ON o.shop_id = s.shop_id
		WHERE
			o.user_id = $2 AND o.status IN ($1, $3)
		ORDER BY
			o.order_date DESC;
	`

	var orders []OrderWithDetailsDB
	if err := r.db.SelectContext(ctx, &orders, query, models.Cooking, userID, models.Completed); err != nil {
		return nil, fmt.Errorf("failed to select active user orders: %w", err)
	}

	return orders, nil
}

// 注文の商品が何なのかとってくる
func (r *orderRepository) FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
	if len(orderIDs) == 0 {
		return make(map[int][]models.ItemDetail), nil
	}

	query, args, err := sqlx.In(`
		SELECT op.order_id, p.item_name, op.quantity
		FROM order_item op
		INNER JOIN items p ON op.item_id = p.item_id
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
	itemsMap := make(map[int][]models.ItemDetail)
	for rows.Next() {
		var orderID int
		var item models.ItemDetail
		if err := rows.Scan(&orderID, &item.ItemName, &item.Quantity); err != nil {
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

//管理者が注文取得
type AdminOrderDBResult struct {
	OrderID       int                `db:"order_id"`
	CustomerEmail sql.NullString     `db:"email"`
	OrderDate     time.Time          `db:"order_date"`
	TotalAmount   float64            `db:"total_amount"`
	Status        models.OrderStatus `db:"status"`
}
func (r *orderRepository) FindActiveShopOrders(ctx context.Context, shopID int) ([]AdminOrderDBResult, error) {
	query := `
		SELECT
			o.order_id, u.email, o.order_date, o.total_amount, o.status
		FROM
			orders o
		LEFT JOIN
			users u ON o.user_id = u.user_id
		WHERE
			o.shop_id = $1 AND o.status IN ($2, $3)
		ORDER BY
			o.order_date ASC
	`
	var orders []AdminOrderDBResult
	if err := r.db.SelectContext(ctx, &orders, query, shopID, models.Cooking, models.Completed); err != nil {
		return nil, fmt.Errorf("failed to select active orders for shop: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
	var order models.Order
	query := `SELECT * FROM orders WHERE order_id = $1 AND shop_id = $2`
	err := r.db.GetContext(ctx, &order, query, orderID, shopID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found or permission denied")
		}
		return nil, err
	}

	return &order, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE order_id = $2 AND shop_id = $3`
	result, err := r.db.ExecContext(ctx, query, newStatus, orderID, shopID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no order was updated, perhaps it was deleted or does not belong to the shop")
	}
	return nil
}

func (r *orderRepository) DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error {
	query := `DELETE FROM orders WHERE order_id = $1 AND shop_id = $2`
	result, err := r.db.ExecContext(ctx, query, orderID, shopID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no order was deleted. order not found or permission denied")
	}

	return nil
}
