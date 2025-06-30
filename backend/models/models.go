package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- UserRole 型と定数の定義 ---
type UserRole int

const (
	UnknownRole  UserRole = iota // 0 (ゼロ値)
	CustomerRole                 // 1
	AdminRole                    // 2
)

func (r UserRole) String() string {
	switch r {
	case CustomerRole:
		return "customer"
	case AdminRole:
		return "admin"
	default:
		return "unknown"
	}
}

// MarshalJSON は、UserRole型をJSONの文字列に変換する
func (r UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON は、JSONの文字列をUserRole型に変換する
func (r *UserRole) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case "customer":
		*r = CustomerRole
	case "admin":
		*r = AdminRole
	default:
		return errors.New("invalid role")
	}
	return nil
}

// --- OrderStatus 型と定数の定義 ---
type OrderStatus int

const (
	UnknownStatus OrderStatus = iota // 0
	Cooking                          // 1 (調理中)
	Completed                        // 2 (調理完了)
	Handed                           // 3 (お渡し済み)
)

func (s OrderStatus) String() string {
	switch s {
	case Cooking:
		return "cooking"
	case Completed:
		return "completed"
	case Handed:
		return "handed"
	default:
		return "unknown"
	}
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *OrderStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "cooking":
		*s = Cooking
	case "completed":
		*s = Completed
	case "handed":
		*s = Handed
	default:
		return errors.New("invalid status")
	}
	return nil
}

type User struct {
	UserID    int       `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type JwtCustomClaims struct {
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
	ShopID *int     `json:"shop_id,omitempty"` // 管理者の場合のみ、担当店舗IDを設定
	jwt.RegisteredClaims
}

func (c *JwtCustomClaims) Valid() error {
	return nil
}

type Shop struct {
	ShopID      int       `json:"shop_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	IsOpen      bool      `json:"is_open"`
	AdminUserID *int      `json:"admin_user_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	OrderID        int            `db:"order_id"`
	UserID         sql.NullInt64  `db:"user_id"` // ゲスト注文ではNULLになる
	ShopID         int            `db:"shop_id"`
	OrderDate      time.Time      `db:"order_date"`
	TotalAmount    float64        `db:"total_amount"`
	UserOrderToken sql.NullString `db:"user_order_token"` // ゲスト注文では一時的なトークンが入る
	Status         OrderStatus    `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type Product struct {
	ProductID   int       `json:"product_id" db:"product_id"`
	ProductName string    `json:"product_name" db:"product_name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OrderProduct struct {
	OrderProductID int       `db:"order_product_id"`
	OrderID        int       `db:"order_id"`
	ProductID      int       `db:"product_id"`
	Quantity       int       `db:"quantity"`
	PriceAtOrder   float64   `db:"price_at_order"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
