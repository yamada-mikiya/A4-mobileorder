package models

import "time"

type User struct {
	UserID uint
	Email string
	CreatedAt time.Time
	UpdatedAt time.Time
}