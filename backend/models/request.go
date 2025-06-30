package models

type CreateOrderRequest struct {
	Products []OrderProductRequest `json:"products"`
}
type OrderProductRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type AuthenticateRequest struct {
	Email          string `json:"email"`
	UserOrderToken string `json:"user_order_token,omitempty"`
}
