package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type ItemRepository interface {
	ValidateAndGetItemsForShop(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error)
	GetItemList(shopID int) ([]models.ItemListResponse, error)
}

type itemRepository struct {
	db *sqlx.DB
}

func NewItemRepository(db *sqlx.DB) ItemRepository {
	return &itemRepository{db}
}

func (r *itemRepository) ValidateAndGetItemsForShop(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error) {
	//商品IDで商品情報取得
	itemMap := make(map[int]models.Item)

	if len(itemIDs) == 0 {
		return itemMap, nil
	}

	const baseQuery = `
		SELECT
			p.item_id,
			p.item_name,
			p.price,
			p.is_available
		FROM
			items p
		INNER JOIN
			shop_item sp ON p.item_id = sp.item_id
		WHERE
			sp.shop_id = ? AND p.item_id IN (?)
	`

	query, args, err := sqlx.In(baseQuery, shopID, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query for item validation: %w", err)
	}
	query = r.db.Rebind(query)

	var items []models.Item
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("failed to select items for shop: %w", err)
	}

	if len(items) != len(itemIDs) {
		return nil, errors.New("one or more items do not belong to the specified shop")
	}

	for _, p := range items {
		itemMap[p.ItemID] = p
	}

	return itemMap, nil

}

func (r *itemRepository) GetItemList(shopID int) ([]models.ItemListResponse, error) {
	//TODO
	return nil, nil
}
