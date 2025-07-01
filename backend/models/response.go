package models

type CreateOrderResponse struct {
	OrderID         int    `json:"order_id" example:"6"`
	UserOrderToken string `json:"user_order_token" example:"15ff4999-2cfd-41f3-b744-926e7c5c7a0"`
	Message         string `json:"message" example:"Order created successfully as a guest. Please sign up to claim this order."`
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

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNiwicm9sZSI6ImN1c3RvbWVyIiwiZXhwIjoxNzUxNTIwMjk1LCJpYXQiOjE3NTEyNjEwOTV9.oItkz3SDGGK0eQSP6BBq-SF3nWLk7Q-ITD1J6UrXeUE"`
}

