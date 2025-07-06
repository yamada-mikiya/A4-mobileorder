package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ShopRepository interface {
	FindshopIDByAdminID(ctx context.Context, userID int) (int, error)
}

type shopRepository struct {
	db *sqlx.DB
}

func NewShopRepository(db *sqlx.DB) ShopRepository {
	return &shopRepository{db}
}

func (r *shopRepository) FindshopIDByAdminID(ctx context.Context, userID int) (int, error) {
	var shopIDs []int
	query := `
		SELECT s.shop_id FROM shops s
		INNER JOIN shop_staff ss ON s.shop_id = ss.shop_id
		WHERE ss.user_id = $1
	`
	err := r.db.SelectContext(ctx, &shopIDs, query, userID)
	if err != nil {
		return 0, err
	}

	switch len(shopIDs) {
	case 0:
		return 0, errors.New("shop not found for the given admin user")
	case 1:
		return shopIDs[0], nil
	default:
		return 0, fmt.Errorf("data inconsistency: user_id %d is associated with multiple shops", userID)
	}
}
