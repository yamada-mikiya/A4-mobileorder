package models

type CreateOrderResponse struct {
	OrderID         int    `json:"order_id"`
	UserOrderToken string `json:"user_order_token"`
	Message         string `json:"message"`
}

type UserResponse struct {
	UserID int   `json:"user_id"`
	Email  string `json:"email"`
	Role string `json:"role"`
}