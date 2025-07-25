package repositories

import (
	"database/sql"
	"errors"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type ShopRepository interface {
	GetShopByAdminID(adminID int) (models.Shop, error)
}

type shopRepository struct {
	db *sqlx.DB
}

func NewShopRepository(db *sqlx.DB) ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) GetShopByAdminID(adminID int) (models.Shop, error) {
	shop := models.Shop{}

	row := r.db.QueryRow(
		`SELECT shop_id, name, description, location, is_open, admin_user_id
         FROM shops
         WHERE admin_user_id = $1`,
		adminID,
	)

	if err := row.Scan(
		&shop.ShopID,
		&shop.Name,
		&shop.Description,
		&shop.Location,
		&shop.IsOpen,
		&shop.AdminUserID,
	); err != nil {
		if err == sql.ErrNoRows {
			return models.Shop{}, errors.New("shop not found for the given admin user")
		}
		return models.Shop{}, err
	}

	return shop, nil
}
