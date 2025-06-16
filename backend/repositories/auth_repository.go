package repositories

import (
	"database/sql"

	"github.com/A4-dev-team/mobileorder.git/models"
)

type IAuthRepository interface {
	CreateUser(user *models.User) error
}

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) IAuthRepository {
	return &AuthRepository{db}
}

func (r *AuthRepository) CreateUser(user *models.User) error {
	//TODO
	return nil
}
