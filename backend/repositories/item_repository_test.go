package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
)

// Test constants for item repository tests (prefixed to avoid conflicts)
const (
	itemTestShopID1 = 1
	itemTestShopID2 = 2
	itemTestItemID1 = 1
	itemTestItemID2 = 2
	itemTestItemID3 = 3
	itemTestItemID4 = 4
	itemTestPrice1  = 100.0
	itemTestPrice2  = 200.0
	itemTestPrice3  = 300.0
	itemTestPrice4  = 400.0
)

func createItemTestShop(shopID int, name string) models.Shop {
	return models.Shop{
		ShopID: shopID,
		Name:   name,
	}
}

func createItemTestItem(itemID int, itemName string, price float64) models.Item {
	return models.Item{
		ItemID:   itemID,
		ItemName: itemName,
		Price:    price,
	}
}

func createItemTestShopItem(shopID, itemID int) models.ShopItem {
	return models.ShopItem{
		ShopID: shopID,
		ItemID: itemID,
	}
}

func setupItemRepositoryTestData(t *testing.T, tx *sqlx.Tx) {
	t.Helper()

	// Insert shops
	shops := []models.Shop{
		createItemTestShop(itemTestShopID1, "Shop A"),
		createItemTestShop(itemTestShopID2, "Shop B"),
	}
	for _, shop := range shops {
		_, err := tx.NamedExec(`INSERT INTO shops (shop_id, name) VALUES (:shop_id, :name)`, shop)
		if err != nil {
			t.Fatalf("failed to insert shop: %v", err)
		}
	}

	// Insert items
	items := []models.Item{
		createItemTestItem(itemTestItemID1, "Item 1", itemTestPrice1),
		createItemTestItem(itemTestItemID2, "Item 2", itemTestPrice2),
		createItemTestItem(itemTestItemID3, "Item 3", itemTestPrice3),
		createItemTestItem(itemTestItemID4, "Item 4", itemTestPrice4),
	}
	for _, item := range items {
		_, err := tx.NamedExec(`INSERT INTO items (item_id, item_name, price) VALUES (:item_id, :item_name, :price)`, item)
		if err != nil {
			t.Fatalf("failed to insert item: %v", err)
		}
	}

	// Insert shop-item relationships
	shopItems := []models.ShopItem{
		createItemTestShopItem(itemTestShopID1, itemTestItemID1),
		createItemTestShopItem(itemTestShopID1, itemTestItemID2),
		createItemTestShopItem(itemTestShopID2, itemTestItemID3),
	}
	for _, si := range shopItems {
		_, err := tx.NamedExec(`INSERT INTO shop_item (shop_id, item_id) VALUES (:shop_id, :item_id)`, si)
		if err != nil {
			t.Fatalf("failed to insert shop_item: %v", err)
		}
	}
}

func assertItemNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("予期せぬエラーが発生しました: %v", err)
	}
}

func assertItemAppError(t *testing.T, err error, expectedErrCode apperrors.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("期待したエラーコード '%s' が発生しませんでした", expectedErrCode)
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("期待したエラーの型(*apperrors.AppError)と異なります: %T", err)
	}

	if expectedErrCode != appErr.ErrCode {
		t.Fatalf("期待したエラーコード '%s', 実際のエラーコード '%s'", expectedErrCode, appErr.ErrCode)
	}
}

func assertItemMapsEqual(t *testing.T, expected, actual map[int]models.Item) {
	t.Helper()
	opts := cmpopts.IgnoreFields(models.Item{}, "Description", "IsAvailable", "CreatedAt", "UpdatedAt")
	if diff := cmp.Diff(expected, actual, opts); diff != "" {
		t.Errorf("item map mismatch (-want +got):\n%s", diff)
	}
}

func TestItemRepository_ValidateAndGetItemsForShop(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	t.Cleanup(func() { db.Close() })

	// Define expected items for test cases
	expectedItem1 := createItemTestItem(itemTestItemID1, "Item 1", itemTestPrice1)
	expectedItem2 := createItemTestItem(itemTestItemID2, "Item 2", itemTestPrice2)

	tests := []struct {
		name          string
		shopID        int
		itemIDs       []int
		expectedMap   map[int]models.Item
		expectErrCode apperrors.ErrCode
	}{
		{
			name:    "正常系: 店舗に属する複数の有効な商品を取得",
			shopID:  itemTestShopID1,
			itemIDs: []int{itemTestItemID1, itemTestItemID2},
			expectedMap: map[int]models.Item{
				itemTestItemID1: expectedItem1,
				itemTestItemID2: expectedItem2,
			},
			expectErrCode: "",
		},
		{
			name:          "正常系: 商品IDリストが空",
			shopID:        itemTestShopID1,
			itemIDs:       []int{},
			expectedMap:   map[int]models.Item{},
			expectErrCode: "",
		},
		{
			name:          "異常系: 存在しない商品IDが含まれている",
			shopID:        itemTestShopID1,
			itemIDs:       []int{itemTestItemID1, 999},
			expectedMap:   nil,
			expectErrCode: apperrors.BadParam,
		},
		{
			name:          "異常系: 他の店舗の商品IDが含まれている",
			shopID:        itemTestShopID1,
			itemIDs:       []int{itemTestItemID1, itemTestItemID3}, // itemTestItemID3 belongs to shop 2
			expectedMap:   nil,
			expectErrCode: apperrors.BadParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer func() {
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Logf("transaction rollback failed: %v", err)
				}
			}()

			setupItemRepositoryTestData(t, tx)

			repo := repositories.NewItemRepository(tx)

			gotMap, err := repo.ValidateAndGetItemsForShop(ctx, tt.shopID, tt.itemIDs)

			if tt.expectErrCode != "" {
				assertItemAppError(t, err, tt.expectErrCode)
			} else {
				assertItemNoError(t, err)
				assertItemMapsEqual(t, tt.expectedMap, gotMap)
			}
		})
	}
}
