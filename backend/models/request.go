package models

type CreateOrderRequest struct {
	Products []OrderProductRequest `json:"products"`
}
type OrderProductRequest struct {
	ProductID int `json:"product_id" example:"1"`
	Quantity  int `json:"quantity" example:"2"`
}

type AuthenticateRequest struct {
	Email          string `json:"email" example:"new.user@example.com"`
	UserOrderToken string `json:"user_order_token,omitempty" example:"15ff4999-2cfd-41f3-b744-926e7c5c7a0e"`
}
