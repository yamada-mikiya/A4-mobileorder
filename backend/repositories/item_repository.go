package repositories

import (
	"context"
	"errors"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/jmoiron/sqlx"
)

type ItemRepository interface {
	ValidateAndGetItemsForShop(ctx context.Context, dbtx DBTX, shopID int, itemIDs []int) (map[int]models.Item, error)
	GetItemList(dbtx DBTX, shopID int) ([]models.ItemListResponse, error)
	UpdateItemAvailability(ctx context.Context, dbtx DBTX, itemID int, isAvailable bool) error
}

type itemRepository struct {
	// ステートレスリポジトリ - DBTXは各メソッドで受け取る
}

func NewItemRepository() ItemRepository {
	return &itemRepository{}
}

func (r *itemRepository) ValidateAndGetItemsForShop(ctx context.Context, dbtx DBTX, shopID int, itemIDs []int) (map[int]models.Item, error) {
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
	query = dbtx.Rebind(query)

	var items []models.Item
	if err := dbtx.SelectContext(ctx, &items, query, args...); err != nil {
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

func (r *itemRepository) GetItemList(dbtx DBTX, shopID int) ([]models.ItemListResponse, error) {
	query := `
		SELECT i.item_id, i.item_name, i.description, i.price, i.is_available
		FROM items i
		INNER JOIN shop_item si ON i.item_id = si.item_id
		WHERE si.shop_id = $1
		ORDER BY i.item_id
	`

	var items []models.Item
	if err := dbtx.SelectContext(context.Background(), &items, query, shopID); err != nil {
		return nil, apperrors.GetDataFailed.Wrap(err, "商品一覧の取得に失敗しました。")
	}

	var response []models.ItemListResponse
	for _, item := range items {
		itemResponse := models.ItemListResponse{
			ItemID:      item.ItemID,
			ItemName:    item.ItemName,
			Description: item.Description,
			Price:       item.Price,
			IsAvailable: item.IsAvailable,
		}

		response = append(response, itemResponse)
	}

	return response, nil
}

// UpdateItemAvailability は商品の販売状態を更新します
func (r *itemRepository) UpdateItemAvailability(ctx context.Context, dbtx DBTX, itemID int, isAvailable bool) error {
	query := `UPDATE items SET is_available = $1, updated_at = NOW() WHERE item_id = $2`

	result, err := dbtx.ExecContext(ctx, query, isAvailable, itemID)
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "商品の販売状態更新に失敗しました。")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.UpdateDataFailed.Wrap(err, "更新結果の確認に失敗しました。")
	}

	if rowsAffected == 0 {
		return apperrors.NoData.Wrap(nil, "指定された商品が見つかりませんでした。")
	}

	return nil
}
