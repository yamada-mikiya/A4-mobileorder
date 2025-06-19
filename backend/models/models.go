package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	UserID    int      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserResponse struct {
	UserID int   `json:"user_id"`
	Email  string `json:"email"`
}

type JwtCustomClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`              // "customer" or "admin"
	ShopID *int   `json:"shop_id,omitempty"` // 管理者の場合のみ、担当店舗IDを設定
	jwt.RegisteredClaims
}

type Shop struct {
	ShopID        int       `json:"shop_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Location      string    `json:"location"`
	IsOpen        bool      `json:"is_open"`
	AdminUserID   *int      `json:"admin_user_id,omitempty"` // 管理者がいない場合(NULL)もあるため、ポインタ型にする
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}