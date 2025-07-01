package models

import "time"

//注文作成レスポンス
type CreateOrderResponse struct {
	OrderID        int    `json:"order_id"`
	UserOrderToken string `json:"user_order_token"`
	Message        string `json:"message"`
}

//ユーザー作成レスポンス
type UserResponse struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

//注文一覧レスポンス
type OrderListResponse struct {
	OrderID      int         `json:"order_id"`  // 注文ID（=注文番号として使う）
	ShopName     string      `json:"shop_name"` // 店舗名
	OrderDate    time.Time   `json:"order_date"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status"`        // "cooking" or "completed"
	WaitingCount int         `json:"waiting_count"` // 待ち人数
	Items        []OrderItem `json:"items"`         // 注文した商品の簡易リスト
}

type OrderItem struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

//注文のステータスと待ち人数表示レスポンス
type OrderStatusResponse struct {
	OrderID      int    `json:"order_id"`
	Status       string `json:"status"`
	WaitingCount int    `json:"waiting_count"`
}