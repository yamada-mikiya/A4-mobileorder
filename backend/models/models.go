package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
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
		return apperrors.ValidationFailed.Wrap(nil, "不正なロール値です: "+s)
	}
	return nil
}

// -----------------定義終わり--------------------------

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
		return apperrors.ValidationFailed.Wrap(nil, "不正なステータス値です: "+str)
	}
	return nil
}

// ---------------定義終わり----------------

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
	ShopID      int       `json:"shop_id" db:"shop_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Location    string    `json:"location" db:"location"`
	IsOpen      bool      `json:"is_open" db:"is_open"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Order struct {
	OrderID         int            `db:"order_id"`
	UserID          sql.NullInt64  `db:"user_id"` // ゲスト注文ではNULLになる
	ShopID          int            `db:"shop_id"`
	OrderDate       time.Time      `db:"order_date"`
	TotalAmount     int            `db:"total_amount"`
	GuestOrderToken sql.NullString `db:"guest_order_token"` // ゲスト注文では一時的なトークンが入る
	Status          OrderStatus    `db:"status"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

type Item struct {
	ItemID      int       `json:"item_id" db:"item_id"`
	ItemName    string    `json:"item_name" db:"item_name"`
	Description string    `json:"description" db:"description"`
	Price       int       `json:"price" db:"price"`
	IsAvailable bool      `json:"is_available" db:"is_available"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	OrderItemID  int       `db:"order_item_id"`
	OrderID      int       `db:"order_id"`
	ItemID       int       `db:"item_id"`
	Quantity     int       `db:"quantity"`
	PriceAtOrder int       `db:"price_at_order"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type ShopItem struct {
	ShopID    int       `db:"shop_id"`
	ItemID    int       `db:"item_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
