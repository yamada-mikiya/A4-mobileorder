package models

import "time"

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}
type OrderItemRequest struct {
	ItemID   int `json:"item_id" validate:"required,min=1" example:"1"`
	Quantity int `json:"quantity" validate:"required,min=1" example:"2"`
}

type AuthenticateRequest struct {
	Email           string `json:"email" validate:"required,email" example:"new.user@example.com"`
	GuestOrderToken string `json:"guest_order_token,omitempty" validate:"omitempty,uuid4" example:"15ff4999-2cfd-41f3-b744-926e7c5c7a0e"`
}

// レート制限用の構造体
type LoginAttempt struct {
	Email     string     `json:"email"`
	Attempts  int        `json:"attempts"`
	LastTry   time.Time  `json:"last_try"`
	BlockedAt *time.Time `json:"blocked_at,omitempty"`
}
