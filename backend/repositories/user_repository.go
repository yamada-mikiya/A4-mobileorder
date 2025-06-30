package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	// 新規ユーザーのデフォルトロールは'customer'に設定
	user.Role = models.CustomerRole
	query := `INSERT INTO users (email, role) VALUES ($1, $2) RETURNING user_id, created_at, updated_at`
	err := r.db.QueryRowxContext(ctx, query, user.Email, user.Role).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return errors.New("email already exists")
		}
		return err
	}
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	user := models.User{}
	query := "SELECT user_id, email, role, created_at, updated_at FROM users WHERE email = $1"
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

