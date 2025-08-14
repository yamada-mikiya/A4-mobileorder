package repositories

import (
	"context"
	"errors"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type ItemRepository interface {
	ValidateAndGetItemsForShop(ctx context.Context, shopID int, itemIDs []int) (map[int]models.Item, error)
	GetItemList(shopID int) ([]models.ItemListResponse, error)
}

type itemRepository struct {
	db DBTX
}

func NewItemRepository(db DBTX) ItemRepository {
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
			i.item_id,
			i.item_name,
			i.price,
			i.is_available
		FROM
			items i
		INNER JOIN
			shop_item si ON i.item_id = si.item_id
		WHERE
			si.shop_id = ? AND i.item_id IN (?)
	`

	query, args, err := sqlx.In(baseQuery, shopID, itemIDs)
	if err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "データベースクエリの構築に失敗しました。")
	}
	query = r.db.Rebind(query)

	var items []models.Item
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "店舗の所属商品情報の取得に失敗しました。")
	}

	if len(items) != len(itemIDs) {
		return nil, apperrors.BadParam.Wrap(errors.New("invalid item id requested"), "リクエストに、存在しないか店舗に属さない商品が含まれています。")
	}

	for _, i := range items {
		itemMap[i.ItemID] = i
	}

	return itemMap, nil

}

func (r *itemRepository) GetItemList(shopID int) ([]models.ItemListResponse, error) {
	//TODO
	return nil, nil
}
