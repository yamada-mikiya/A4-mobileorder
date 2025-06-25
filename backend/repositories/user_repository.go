package repositories

import (
	"database/sql"
	"errors"

	"github.com/A4-dev-team/mobileorder.git/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) CreateUser(user *models.User) error {

	err := r.db.QueryRow(
		"INSERT INTO users (email) VALUES ($1) RETURNING user_id",
		user.Email,
	).Scan(&user.UserID)

	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserByEmail(email string) (models.User, error) {
	user := models.User{}

	row := r.db.QueryRow(
		"SELECT user_id, email, role FROM users WHERE email = $1",
		email,
	)
	if err := row.Scan(&user.UserID, &user.Email, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}
