package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
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
	FindShopOrdersByStatuses(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]AdminOrderDBResult, error)
	FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error
	DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error
}

type orderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	orderQuery := `
		INSERT INTO orders (user_id, shop_id, order_date, total_amount, guest_order_token, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING order_id, created_at, updated_at
	`
	err := r.db.QueryRowxContext(
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
		return apperrors.InsertDataFailed.Wrap(err, "注文の作成に失敗しました。")
	}

	stmt, err := r.db.PreparexContext(ctx, "INSERT INTO order_item (order_id, item_id, quantity, price_at_order) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return apperrors.InsertDataFailed.Wrap(err, "注文商品登録の準備に失敗しました。")
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err = stmt.ExecContext(ctx, order.OrderID, item.ItemID, item.Quantity, item.PriceAtOrder); err != nil {
			return apperrors.InsertDataFailed.Wrap(err, "注文商品の登録に失敗しました。")
		}
	}

	return nil
}

func (r *orderRepository) UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error {
	// user_id が NULL の場合のみ更新（まだユーザーに紐付けられていない注文のみ）
	query := "UPDATE orders SET user_id = $1 WHERE guest_order_token = $2 AND user_id IS NULL"
	result, err := r.db.ExecContext(ctx, query, userID, guestToken)
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "ゲスト注文のユーザー紐付けに失敗しました。")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "更新結果の取得に失敗しました。")
	}
	if rowsAffected == 0 {
		// より詳細なエラーチェックのため、注文が存在するかチェック
		var existingUserID sql.NullInt64
		checkQuery := "SELECT user_id FROM orders WHERE guest_order_token = $1"
		err := r.db.QueryRowxContext(ctx, checkQuery, guestToken).Scan(&existingUserID)

		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.NoData.Wrap(nil, "指定されたゲスト注文トークンが見つかりませんでした。")
			}
			return apperrors.GetDataFailed.Wrap(err, "ゲスト注文の確認に失敗しました。")
		}

		if existingUserID.Valid {
			return apperrors.Conflict.Wrap(nil, "この注文は既に他のユーザーアカウントに紐付けられています。")
		}

		return apperrors.NoData.Wrap(nil, "指定されたゲスト注文は見つかりませんでした。")
	}
	return nil
}

// ユーザーの注文情報をとってくる。
type OrderWithDetailsDB struct {
	OrderID      int                `db:"order_id"`
	ShopName     string             `db:"shop_name"`
	Location     string             `db:"location"`
	OrderDate    time.Time          `db:"order_date"`
	TotalAmount  int                `db:"total_amount"`
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
			o.status IN ($1, $2) AND o.user_id = $3
		ORDER BY
			o.order_date DESC;
	`

	var orders []OrderWithDetailsDB
	if err := r.db.SelectContext(ctx, &orders, query, models.Cooking, models.Completed, userID); err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "アクティブな注文履歴の取得に失敗しました。")
	}

	return orders, nil
}

// 注文IDに対応する商品をとってくる
func (r *orderRepository) FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
	if len(orderIDs) == 0 {
		return make(map[int][]models.ItemDetail), nil
	}

	query, args, err := sqlx.In(`
		SELECT oi.order_id, i.item_name, oi.quantity
		FROM order_item oi
		INNER JOIN items i ON oi.item_id = i.item_id
		WHERE oi.order_id IN (?)
	`, orderIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "データベースクエリの実行に失敗しました。")
	}
	defer rows.Close()
	itemsMap := make(map[int][]models.ItemDetail)
	for rows.Next() {
		var orderID int
		var item models.ItemDetail
		if err := rows.Scan(&orderID, &item.ItemName, &item.Quantity); err != nil {
			return nil, apperrors.GetDataFailed.Wrap(err, "注文商品データの読み取りに失敗しました。")
		}
		itemsMap[orderID] = append(itemsMap[orderID], item)
	}

	return itemsMap, nil
}

func (r *orderRepository) FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	var order models.Order
	// アクティブな注文（cooking, completed）のみを取得
	query := "SELECT * FROM orders WHERE order_id = $1 AND user_id = $2 AND status IN ($3, $4)"
	if err := r.db.GetContext(ctx, &order, query, orderID, userID, models.Cooking, models.Completed); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NoData.Wrap(err, "注文が見つからないか、アクセス権がありません。")
		}
		return nil, apperrors.GetDataFailed.Wrap(err, "注文情報の取得に失敗しました。")
	}
	return &order, nil
}

func (r *orderRepository) CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE shop_id = $1 AND status = $2 AND order_date < $3`
	if err := r.db.GetContext(ctx, &count, query, shopID, models.Cooking, orderDate); err != nil {
		return 0, apperrors.GetDataFailed.Wrap(err, "待ち人数の取得に失敗しました。")
	}
	return count, nil
}

// 管理者が注文取得
type AdminOrderDBResult struct {
	OrderID       int                `db:"order_id"`
	CustomerEmail sql.NullString     `db:"email"`
	OrderDate     time.Time          `db:"order_date"`
	TotalAmount   int                `db:"total_amount"` // 円単位の整数
	Status        models.OrderStatus `db:"status"`
}

func (r *orderRepository) FindShopOrdersByStatuses(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]AdminOrderDBResult, error) {
	if len(statuses) == 0 {
		return []AdminOrderDBResult{}, nil
	}
	query, args, err := sqlx.In(`
		SELECT
			o.order_id, u.email, o.order_date, o.total_amount, o.status
		FROM
			orders o
		LEFT JOIN
			users u ON o.user_id = u.user_id
		WHERE
			o.shop_id = ? AND o.status IN (?)
		ORDER BY
			o.order_date ASC
	`, shopID, statuses)
	if err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "データベースクエリの構築に失敗しました。")
	}
	query = r.db.Rebind(query)

	var orders []AdminOrderDBResult
	if err := r.db.SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "店舗の注文情報取得に失敗しました。")
	}
	return orders, nil
}

func (r *orderRepository) FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
	var order models.Order
	query := `SELECT * FROM orders WHERE order_id = $1 AND shop_id = $2`
	err := r.db.GetContext(ctx, &order, query, orderID, shopID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NoData.Wrap(err, "注文が見つからないか、この店舗の管轄外です。")
		}
		return nil, apperrors.GetDataFailed.Wrap(err, "注文情報の取得に失敗しました。")
	}

	return &order, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE order_id = $2 AND shop_id = $3`
	result, err := r.db.ExecContext(ctx, query, newStatus, orderID, shopID)
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "注文ステータスの更新に失敗しました。")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "更新結果の取得に失敗しました。")
	}
	if rowsAffected == 0 {
		return apperrors.NoData.Wrap(nil, "更新対象の注文が見つからないか、管轄外です。")
	}
	return nil
}

func (r *orderRepository) DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error {
	query := `DELETE FROM orders WHERE order_id = $1 AND shop_id = $2`
	result, err := r.db.ExecContext(ctx, query, orderID, shopID)
	if err != nil {
		return apperrors.DeleteDataFailed.Wrap(err, "注文の削除に失敗しました。")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DeleteDataFailed.Wrap(err, "削除結果の取得に失敗しました。")
	}

	if rowsAffected == 0 {
		return apperrors.NoData.Wrap(nil, "削除対象の注文が見つからないか、管轄外です。")
	}

	return nil
}
