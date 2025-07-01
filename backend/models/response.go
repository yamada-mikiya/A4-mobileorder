package models

type CreateOrderResponse struct {
	OrderID         int    `json:"order_id"`
	UserOrderToken string `json:"user_order_token"`
	Message         string `json:"message"`
}

type UserResponse struct {
	UserID int   `json:"user_id" example:"16"`
	Email  string `json:"email" example:"new.user@example.com"`
	Role string `json:"role" exapmple:"customer"`
}

type SignUpResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNiwicm9sZSI6ImN1c3RvbWVyIiwiZXhwIjoxNzUxNTIwMjk1LCJpYXQiOjE3NTEyNjEwOTV9.oItkz3SDGGK0eQSP6BBq-SF3nWLk7Q-ITD1J6UrXeUE"`
	User UserResponse `json:"user"`	
}
