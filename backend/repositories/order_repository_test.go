package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
)

// テスト定数
const (
	testUserID    = 1
	testShopID    = 1
	testUserID2   = 2
	testUserID3   = 3
	testShopID2   = 2
	testItemID1   = 1
	testItemID2   = 2
	testPrice1    = 100.0
	testPrice2    = 200.0
	testAmount1   = 500.0
	testAmount2   = 1000.0
	testQuantity1 = 1
	testQuantity2 = 2
	testQuantity3 = 3

	// 新しいテスト用定数
	testUserID1        = 1
	testShopID1        = 1
	testOrderID1       = 1
	testOrderID2       = 2
	testOrderID3       = 3
	testOrderID4       = 4
	testOrderID5       = 5
	testTotalAmount1   = 1000.0
	nonExistentOrderID = 9999
)

func newTestUser(id int) models.User {
	return models.User{
		UserID: id,
		Email:  fmt.Sprintf("user%d@test.com", id),
	}
}

func newTestShop(id int) models.Shop {
	return models.Shop{
		ShopID: id,
		Name:   fmt.Sprintf("Test Shop %d", id),
	}
}

func newTestShopWithLocation(id int, name, location string) models.Shop {
	return models.Shop{
		ShopID:   id,
		Name:     name,
		Location: location,
	}
}

func newTestItem(id int, name string, price float64) models.Item {
	return models.Item{
		ItemID:   id,
		ItemName: name,
		Price:    price,
	}
}

func newTestItems() []models.Item {
	return []models.Item{
		newTestItem(testItemID1, "Item A", testPrice1),
		newTestItem(testItemID2, "Item B", testPrice2),
	}
}

func newTestOrderItem(orderID, itemID, quantity int, priceAtOrder float64) models.OrderItem {
	return models.OrderItem{
		OrderID:      orderID,
		ItemID:       itemID,
		Quantity:     quantity,
		PriceAtOrder: priceAtOrder,
	}
}

func newTestOrder(userID, shopID int, totalAmount float64, status models.OrderStatus) *models.Order {
	return &models.Order{
		UserID:      sql.NullInt64{Int64: int64(userID), Valid: true},
		ShopID:      shopID,
		TotalAmount: totalAmount,
		Status:      status,
	}
}

// createTestUser - テスト用ユーザーをDBに作成するヘルパー関数
func createTestUser(t *testing.T, tx *sqlx.Tx, userID int, email string) {
	t.Helper()
	query := `INSERT INTO users (user_id, email) VALUES ($1, $2)`
	_, err := tx.Exec(query, userID, email)
	if err != nil {
		t.Fatalf("テスト用ユーザーの作成に失敗しました: %v", err)
	}
}

// createTestShop - テスト用店舗をDBに作成するヘルパー関数
func createTestShop(t *testing.T, tx *sqlx.Tx, shopID int, name string) {
	t.Helper()
	query := `INSERT INTO shops (shop_id, name) VALUES ($1, $2)`
	_, err := tx.Exec(query, shopID, name)
	if err != nil {
		t.Fatalf("テスト用店舗の作成に失敗しました: %v", err)
	}
}

// createTestOrder - テスト用注文をDBに作成するヘルパー関数
func createTestOrder(t *testing.T, tx *sqlx.Tx, orderID, userID, shopID int, status models.OrderStatus) {
	t.Helper()
	query := `INSERT INTO orders (order_id, user_id, shop_id, total_amount, status, order_date) 
			  VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := tx.Exec(query, orderID, userID, shopID, testTotalAmount1, status)
	if err != nil {
		t.Fatalf("テスト用注文の作成に失敗しました: %v", err)
	}
}

// createTestOrderWithTime - 指定時刻でテスト用注文をDBに作成するヘルパー関数
func createTestOrderWithTime(t *testing.T, tx *sqlx.Tx, orderID, userID, shopID int, status models.OrderStatus, orderDate time.Time) {
	t.Helper()
	query := `INSERT INTO orders (order_id, user_id, shop_id, total_amount, status, order_date) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(query, orderID, userID, shopID, testTotalAmount1, status, orderDate)
	if err != nil {
		t.Fatalf("テスト用注文の作成に失敗しました: %v", err)
	}
}

func TestOrderRepository_CreateOrder(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	t.Cleanup(func() { db.Close() })

	tests := []struct {
		name              string
		order             *models.Order
		items             []models.OrderItem
		expectedErrCode   apperrors.ErrCode
		expectedItemCount int
	}{
		{
			name:  "正常系: アイテム付きで注文が正常に作成される",
			order: newTestOrder(testUserID, testShopID, testAmount1, models.Cooking),
			items: []models.OrderItem{
				newTestOrderItem(0, testItemID1, testQuantity1, testPrice1),
				newTestOrderItem(0, testItemID2, testQuantity2, testPrice2),
			},
			expectedItemCount: 2,
		},
		{
			name:  "異常系: 存在しないアイテムIDの場合は失敗する",
			order: newTestOrder(testUserID, testShopID, 0, models.Cooking),
			items: []models.OrderItem{
				newTestOrderItem(0, 999, testQuantity1, testPrice1),
			},
			expectedErrCode: apperrors.InsertDataFailed,
		},
		{
			name:              "正常系: アイテムなしで注文が正常に作成される",
			order:             newTestOrder(testUserID, testShopID, 0, models.Cooking),
			items:             []models.OrderItem{},
			expectedItemCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// テスト分離のためのトランザクション設定
			tx := db.MustBegin()
			defer func() {
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Logf("トランザクションのロールバックに失敗しました: %v", err)
				}
			}()

			// テスト前提データの作成
			createTestUser(t, tx, testUserID, fmt.Sprintf("user%d@test.com", testUserID))
			createTestShop(t, tx, testShopID, fmt.Sprintf("Test Shop %d", testShopID))

			// アイテムを直接作成
			items := newTestItems()
			for _, item := range items {
				_, err := tx.NamedExec(`INSERT INTO items (item_id, item_name, price) VALUES (:item_id, :item_name, :price)`, item)
				if err != nil {
					t.Fatalf("アイテムの挿入に失敗しました: %v", err)
				}
			}

			repo := repositories.NewOrderRepository(tx)
			err := repo.CreateOrder(ctx, tt.order, tt.items)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				// 注文が正しく作成されているかチェック
				if tt.order.OrderID == 0 {
					t.Error("OrderIDが設定されていません")
				}

				var orderCount int
				err := tx.Get(&orderCount, "SELECT COUNT(*) FROM orders WHERE order_id = $1", tt.order.OrderID)
				if err != nil {
					t.Fatalf("DBから注文数を取得できませんでした: %v", err)
				}
				if orderCount != 1 {
					t.Errorf("期待した注文レコード数は1ですが、実際は %d でした", orderCount)
				}

				var itemCount int
				err = tx.Get(&itemCount, "SELECT COUNT(*) FROM order_item WHERE order_id = $1", tt.order.OrderID)
				if err != nil {
					t.Fatalf("DBから注文アイテム数を取得できませんでした: %v", err)
				}
				if itemCount != tt.expectedItemCount {
					t.Errorf("期待した注文アイテムレコード数は %d ですが、実際は %d でした", tt.expectedItemCount, itemCount)
				}
			}
		})
	}
}

const (
	testGuestToken1 = "test-guest-token-1"
	testGuestToken2 = "non-existent-token"
)

func TestOrderRepository_UpdateUserIDByGuestToken(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	t.Cleanup(func() { db.Close() })

	tests := []struct {
		name            string
		guestToken      string
		userIDToSet     int
		setup           func(t *testing.T, tx *sqlx.Tx)
		expectedErrCode apperrors.ErrCode
		assertion       func(t *testing.T, tx *sqlx.Tx)
	}{
		{
			name:        "正常系: ゲスト注文のuser_idが正常に更新される",
			guestToken:  testGuestToken1,
			userIDToSet: testUserID,
			setup: func(t *testing.T, tx *sqlx.Tx) {
				// テスト前提データの作成
				createTestUser(t, tx, testUserID, fmt.Sprintf("user%d@test.com", testUserID))
				createTestShop(t, tx, testShopID, fmt.Sprintf("Test Shop %d", testShopID))

				// アイテムを直接作成
				items := newTestItems()
				for _, item := range items {
					_, err := tx.NamedExec(`INSERT INTO items (item_id, item_name, price) VALUES (:item_id, :item_name, :price)`, item)
					if err != nil {
						t.Fatalf("アイテムの挿入に失敗しました: %v", err)
					}
				}

				tx.MustExec(`
					INSERT INTO orders (shop_id, order_date, total_amount, guest_order_token, status)
					VALUES ($1, NOW(), $2, $3, $4)
				`, testShopID, testAmount2, testGuestToken1, models.Cooking)
			},
			assertion: func(t *testing.T, tx *sqlx.Tx) {
				var updatedUserID sql.NullInt64
				err := tx.Get(&updatedUserID, "SELECT user_id FROM orders WHERE guest_order_token = $1", testGuestToken1)
				if err != nil {
					t.Fatalf("更新結果の取得に失敗しました: %v", err)
				}
				if !updatedUserID.Valid || updatedUserID.Int64 != int64(testUserID) {
					t.Errorf("user_idが正しく更新されていません: 取得値 %v, 期待値 %d", updatedUserID, testUserID)
				}
			},
		},
		{
			name:        "異常系: ゲストトークンが存在しない場合は失敗する",
			guestToken:  testGuestToken2,
			userIDToSet: testUserID,
			setup: func(t *testing.T, tx *sqlx.Tx) {
				// テスト前提データの作成
				createTestUser(t, tx, testUserID, fmt.Sprintf("user%d@test.com", testUserID))
				createTestShop(t, tx, testShopID, fmt.Sprintf("Test Shop %d", testShopID))

				// アイテムを直接作成
				items := newTestItems()
				for _, item := range items {
					_, err := tx.NamedExec(`INSERT INTO items (item_id, item_name, price) VALUES (:item_id, :item_name, :price)`, item)
					if err != nil {
						t.Fatalf("アイテムの挿入に失敗しました: %v", err)
					}
				}
			},
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer func() {
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Logf("トランザクションのロールバックに失敗しました: %v", err)
				}
			}()

			if tt.setup != nil {
				tt.setup(t, tx)
			}

			repo := repositories.NewOrderRepository(tx)
			err := repo.UpdateUserIDByGuestToken(ctx, tt.guestToken, tt.userIDToSet)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, tx)
				}
			}
		})
	}
}

func setupActiveUserOrdersData(t *testing.T, tx *sqlx.Tx) {
	t.Helper()

	users := []models.User{
		newTestUser(testUserID),
		newTestUser(testUserID2),
		newTestUser(testUserID3),
	}
	for _, user := range users {
		_, err := tx.NamedExec(`INSERT INTO users (user_id, email) VALUES (:user_id, :email)`, user)
		if err != nil {
			t.Fatalf("ユーザーの挿入に失敗しました: %v", err)
		}
	}

	shops := []models.Shop{
		newTestShopWithLocation(testShopID, "Shop A", "Location A"),
		newTestShopWithLocation(testShopID2, "Shop B", "Location B"),
	}
	for _, shop := range shops {
		_, err := tx.NamedExec(`INSERT INTO shops (shop_id, name, location) VALUES (:shop_id, :name, :location)`, shop)
		if err != nil {
			t.Fatalf("ショップの挿入に失敗しました: %v", err)
		}
	}

	// テスト一貫性のための固定タイムスタンプでテスト注文を挿入
	baseTime := time.Date(2025, 8, 14, 10, 0, 0, 0, time.UTC)
	orders := []struct {
		orderID  int
		userID   int
		shopID   int
		duration time.Duration
		status   models.OrderStatus
	}{
		{1, testUserID, testShopID, 1 * time.Minute, models.Cooking},
		{2, testUserID, testShopID, 2 * time.Minute, models.Cooking},
		{3, testUserID, testShopID, 3 * time.Minute, models.Completed},
		{4, testUserID, testShopID, 4 * time.Minute, models.Handed}, // 非アクティブ
		{5, testUserID2, testShopID, 0 * time.Minute, models.Cooking},
		{6, testUserID, testShopID2, 0 * time.Minute, models.Cooking}, // 異なるショップ
		{7, testUserID3, testShopID, 5 * time.Minute, models.Handed},  // 非アクティブ
	}

	for _, o := range orders {
		_, err := tx.Exec(`
			INSERT INTO orders (order_id, user_id, shop_id, order_date, total_amount, status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, o.orderID, o.userID, o.shopID, baseTime.Add(o.duration), testPrice1, o.status)
		if err != nil {
			t.Fatalf("テストセットアップ用注文の挿入に失敗しました: %v", err)
		}
	}
}

func TestOrderRepository_FindActiveUserOrders(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	t.Cleanup(func() { db.Close() })

	tests := []struct {
		name            string
		userID          int
		expectedErrCode apperrors.ErrCode
		assertion       func(t *testing.T, got []repositories.OrderWithDetailsDB)
	}{
		{
			name:   "正常系: ユーザー1の正しい待機数付きアクティブ注文を返す",
			userID: testUserID,
			assertion: func(t *testing.T, got []repositories.OrderWithDetailsDB) {
				expected := []repositories.OrderWithDetailsDB{
					{OrderID: 3, ShopName: "Shop A", Location: "Location A", Status: models.Completed, WaitingCount: 0},
					{OrderID: 2, ShopName: "Shop A", Location: "Location A", Status: models.Cooking, WaitingCount: 2},
					{OrderID: 1, ShopName: "Shop A", Location: "Location A", Status: models.Cooking, WaitingCount: 1},
					{OrderID: 6, ShopName: "Shop B", Location: "Location B", Status: models.Cooking, WaitingCount: 0},
				}
				opts := []cmp.Option{
					cmpopts.IgnoreFields(repositories.OrderWithDetailsDB{}, "OrderDate", "TotalAmount"),
					cmpopts.EquateEmpty(),
				}
				if diff := cmp.Diff(expected, got, opts...); diff != "" {
					t.Errorf("FindActiveUserOrders の結果が一致しません (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:   "正常系: ユーザーにアクティブな注文がない場合は空のスライスを返す",
			userID: testUserID3,
			assertion: func(t *testing.T, got []repositories.OrderWithDetailsDB) {
				if len(got) != 0 {
					t.Errorf("空のスライスを期待していましたが、%d 個のアイテムを取得しました", len(got))
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer func() {
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Logf("トランザクションのロールバックに失敗しました: %v", err)
				}
			}()

			setupActiveUserOrdersData(t, tx)

			repo := repositories.NewOrderRepository(tx)
			got, err := repo.FindActiveUserOrders(ctx, tt.userID)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, got)
				}
			}
		})
	}
}

func setupItemsByOrderIDsData(t *testing.T, tx *sqlx.Tx) {
	t.Helper()

	user := newTestUser(testUserID)
	shop := newTestShop(testShopID)

	_, err := tx.NamedExec(`INSERT INTO users (user_id, email) VALUES (:user_id, :email)`, user)
	if err != nil {
		t.Fatalf("ユーザーの挿入に失敗しました: %v", err)
	}

	_, err = tx.NamedExec(`INSERT INTO shops (shop_id, name) VALUES (:shop_id, :name)`, shop)
	if err != nil {
		t.Fatalf("ショップの挿入に失敗しました: %v", err)
	}

	items := []models.Item{
		newTestItem(testItemID1, "Item A", testPrice1),
		newTestItem(testItemID2, "Item B", testPrice2),
	}
	for _, item := range items {
		_, err = tx.NamedExec(`INSERT INTO items (item_id, item_name, price) VALUES (:item_id, :item_name, :price)`, item)
		if err != nil {
			t.Fatalf("アイテムの挿入に失敗しました: %v", err)
		}
	}

	// 注文を挿入
	orders := []struct {
		orderID int
		comment string
	}{
		{1, "アイテム付き注文"},
		{2, "異なるアイテム付き注文"},
		{3, "アイテムなし注文"},
	}

	for _, order := range orders {
		_, err := tx.Exec(`
			INSERT INTO orders (order_id, user_id, shop_id, order_date, total_amount, status)
			VALUES ($1, $2, $3, NOW(), $4, $5)
		`, order.orderID, testUserID, testShopID, testPrice1, models.Cooking)
		if err != nil {
			t.Fatalf("テストセットアップ用注文 %d の挿入に失敗しました: %v", order.orderID, err)
		}
	}

	// 注文アイテムを挿入
	orderItems := []models.OrderItem{
		// 注文1
		newTestOrderItem(1, testItemID1, testQuantity1, testPrice1),
		newTestOrderItem(1, testItemID2, testQuantity2, testPrice2),
		// 注文2
		newTestOrderItem(2, testItemID1, testQuantity3, testPrice1),
		// 注文3にはアイテムなし
	}

	for _, oi := range orderItems {
		_, err := tx.NamedExec(`
			INSERT INTO order_item (order_id, item_id, quantity, price_at_order)
			VALUES (:order_id, :item_id, :quantity, :price_at_order)
		`, oi)
		if err != nil {
			t.Fatalf("テストセットアップ用注文アイテムの挿入に失敗しました: %v", err)
		}
	}
}

func TestOrderRepository_FindItemsByOrderIDs(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	t.Cleanup(func() { db.Close() })

	tests := []struct {
		name            string
		orderIDs        []int
		expectedErrCode apperrors.ErrCode
		assertion       func(t *testing.T, got map[int][]models.ItemDetail)
	}{
		{
			name:     "正常系: 複数の注文IDに対してアイテムを返す",
			orderIDs: []int{1, 2},
			assertion: func(t *testing.T, got map[int][]models.ItemDetail) {
				expected := map[int][]models.ItemDetail{
					1: {
						{ItemName: "Item A", Quantity: testQuantity1},
						{ItemName: "Item B", Quantity: testQuantity2},
					},
					2: {
						{ItemName: "Item A", Quantity: testQuantity3},
					},
				}
				opts := cmpopts.SortSlices(func(a, b models.ItemDetail) bool { return a.ItemName < b.ItemName })
				if diff := cmp.Diff(expected, got, opts); diff != "" {
					t.Errorf("FindItemsByOrderIDs の結果が一致しません (-want +got):\n%s", diff)
				}
			},
		},
		{
			name:     "正常系: 注文IDリストが空の場合は空のマップを返す",
			orderIDs: []int{},
			assertion: func(t *testing.T, got map[int][]models.ItemDetail) {
				if len(got) != 0 {
					t.Errorf("空のマップを期待していましたが、%d 個のアイテムを取得しました", len(got))
				}
			},
		},
		{
			name:     "正常系: アイテムがない注文IDは結果に含まれない",
			orderIDs: []int{1, 3}, // 注文3にはアイテムがない
			assertion: func(t *testing.T, got map[int][]models.ItemDetail) {
				expected := map[int][]models.ItemDetail{
					1: {
						{ItemName: "Item A", Quantity: testQuantity1},
						{ItemName: "Item B", Quantity: testQuantity2},
					},
				}
				if _, ok := got[3]; ok {
					t.Error("アイテムがない注文ID 3 は結果に含まれるべきではありません")
				}
				opts := cmpopts.SortSlices(func(a, b models.ItemDetail) bool { return a.ItemName < b.ItemName })
				if diff := cmp.Diff(expected, got, opts); diff != "" {
					t.Errorf("FindItemsByOrderIDs の結果が一致しません (-want +got):\n%s", diff)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt // ループ変数をキャプチャ
		t.Run(tt.name, func(t *testing.T) {
			// テスト分離のためのトランザクション設定
			tx := db.MustBegin()
			defer func() {
				if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Logf("トランザクションのロールバックに失敗しました: %v", err)
				}
			}()

			// テストデータのセットアップ
			setupItemsByOrderIDsData(t, tx)

			// リポジトリ作成とテスト実行
			repo := repositories.NewOrderRepository(tx)
			got, err := repo.FindItemsByOrderIDs(ctx, tt.orderIDs)

			// 結果のアサーション
			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if tt.assertion != nil {
					tt.assertion(t, got)
				}
			}
		})
	}
}

// TestFindOrderByIDAndUser - ユーザーIDと注文IDで注文を取得するテスト
func TestFindOrderByIDAndUser(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		orderID         int
		userID          int
		setup           func(*sqlx.Tx)
		want            *models.Order
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:    "正常に注文を取得できる",
			orderID: testOrderID1,
			userID:  testUserID1,
			setup: func(tx *sqlx.Tx) {
				// 前提データを作成
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
			want: &models.Order{
				OrderID:     testOrderID1,
				UserID:      sql.NullInt64{Int64: int64(testUserID1), Valid: true},
				ShopID:      testShopID1,
				Status:      models.Cooking,
				TotalAmount: testTotalAmount1,
			},
		},
		{
			name:    "異なるユーザーIDでは注文が取得できない",
			orderID: testOrderID1,
			userID:  testUserID2,
			setup: func(tx *sqlx.Tx) {
				// 前提データを作成（testUserID1で注文作成、testUserID2でアクセス試行）
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
			expectedErrCode: apperrors.NoData,
		},
		{
			name:            "存在しない注文IDでは注文が取得できない",
			orderID:         nonExistentOrderID,
			userID:          testUserID1,
			setup:           func(tx *sqlx.Tx) {},
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			got, err := repo.FindOrderByIDAndUser(context.Background(), tt.orderID, tt.userID)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if tt.want != nil {
					// 注文の詳細を検証
					if got == nil {
						t.Error("got order is nil")
						return
					}
					if tt.want.OrderID != got.OrderID {
						t.Errorf("OrderID: expected %d, got %d", tt.want.OrderID, got.OrderID)
					}
					if tt.want.UserID != got.UserID {
						t.Errorf("UserID: expected %v, got %v", tt.want.UserID, got.UserID)
					}
					if tt.want.ShopID != got.ShopID {
						t.Errorf("ShopID: expected %d, got %d", tt.want.ShopID, got.ShopID)
					}
					if tt.want.Status != got.Status {
						t.Errorf("Status: expected %v, got %v", tt.want.Status, got.Status)
					}
					if tt.want.TotalAmount != got.TotalAmount {
						t.Errorf("TotalAmount: expected %f, got %f", tt.want.TotalAmount, got.TotalAmount)
					}
				}
			}
		})
	}
}

// TestCountWaitingOrders - 待ち人数をカウントするテスト
func TestCountWaitingOrders(t *testing.T) {
	db := NewTestDB(t)

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	earlierTime := baseTime.Add(-1 * time.Hour)
	laterTime := baseTime.Add(1 * time.Hour)

	tests := []struct {
		name            string
		shopID          int
		orderDate       time.Time
		setup           func(*sqlx.Tx)
		want            int
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:      "調理中の注文数を正しくカウントできる",
			shopID:    testShopID1,
			orderDate: baseTime,
			setup: func(tx *sqlx.Tx) {
				// 前提データを作成
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestUser(t, tx, testUserID2, fmt.Sprintf("user%d@test.com", testUserID2))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestShop(t, tx, testShopID2, fmt.Sprintf("Test Shop %d", testShopID2))

				// 注文を作成
				createTestOrderWithTime(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking, earlierTime)
				createTestOrderWithTime(t, tx, testOrderID2, testUserID2, testShopID1, models.Cooking, earlierTime.Add(-30*time.Minute))
				createTestOrderWithTime(t, tx, testOrderID3, testUserID1, testShopID1, models.Cooking, laterTime)
				createTestOrderWithTime(t, tx, testOrderID4, testUserID2, testShopID1, models.Completed, earlierTime)
				createTestOrderWithTime(t, tx, testOrderID5, testUserID1, testShopID2, models.Cooking, earlierTime)
			},
			want: 2,
		},
		{
			name:      "該当する注文がない場合は0を返す",
			shopID:    testShopID1,
			orderDate: baseTime,
			setup:     func(tx *sqlx.Tx) {},
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			got, err := repo.CountWaitingOrders(context.Background(), tt.shopID, tt.orderDate)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if got != tt.want {
					t.Errorf("expected count %d, got %d", tt.want, got)
				}
			}
		})
	}
}

// TestFindShopOrdersByStatuses - ステータス別店舗注文取得テスト
func TestFindShopOrdersByStatuses(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		shopID          int
		statuses        []models.OrderStatus
		setup           func(*sqlx.Tx)
		want            []repositories.AdminOrderDBResult
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:     "指定ステータスの注文を正しく取得できる",
			shopID:   testShopID1,
			statuses: []models.OrderStatus{models.Completed, models.Cooking},
			setup: func(tx *sqlx.Tx) {
				// 前提データを作成
				createTestUser(t, tx, testUserID1, "test1@example.com")
				createTestUser(t, tx, testUserID2, "test2@example.com")
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestShop(t, tx, testShopID2, fmt.Sprintf("Test Shop %d", testShopID2))

				// 注文を作成
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Completed)
				createTestOrder(t, tx, testOrderID2, testUserID2, testShopID1, models.Cooking)
				createTestOrder(t, tx, testOrderID3, testUserID1, testShopID1, models.Handed)
				createTestOrder(t, tx, testOrderID4, testUserID1, testShopID2, models.Completed)
			},
			want: []repositories.AdminOrderDBResult{
				{
					OrderID:       testOrderID1,
					CustomerEmail: sql.NullString{String: "test1@example.com", Valid: true},
					TotalAmount:   testTotalAmount1,
					Status:        models.Completed,
				},
				{
					OrderID:       testOrderID2,
					CustomerEmail: sql.NullString{String: "test2@example.com", Valid: true},
					TotalAmount:   testTotalAmount1,
					Status:        models.Cooking,
				},
			},
		},
		{
			name:     "空のステータス配列では空の結果を返す",
			shopID:   testShopID1,
			statuses: []models.OrderStatus{},
			setup:    func(tx *sqlx.Tx) {},
			want:     []repositories.AdminOrderDBResult{},
		},
		{
			name:     "該当する注文がない場合は空の結果を返す",
			shopID:   testShopID1,
			statuses: []models.OrderStatus{models.Completed},
			setup:    func(tx *sqlx.Tx) {},
			want:     []repositories.AdminOrderDBResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			got, err := repo.FindShopOrdersByStatuses(context.Background(), tt.shopID, tt.statuses)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				if len(got) != len(tt.want) {
					t.Errorf("expected %d orders, got %d", len(tt.want), len(got))
					return
				}

				for i, want := range tt.want {
					if i >= len(got) {
						t.Errorf("missing order at index %d", i)
						continue
					}

					if got[i].OrderID != want.OrderID {
						t.Errorf("order %d: expected OrderID %d, got %d", i, want.OrderID, got[i].OrderID)
					}
					if got[i].CustomerEmail != want.CustomerEmail {
						t.Errorf("order %d: expected CustomerEmail %v, got %v", i, want.CustomerEmail, got[i].CustomerEmail)
					}
					if got[i].TotalAmount != want.TotalAmount {
						t.Errorf("order %d: expected TotalAmount %f, got %f", i, want.TotalAmount, got[i].TotalAmount)
					}
					if got[i].Status != want.Status {
						t.Errorf("order %d: expected Status %v, got %v", i, want.Status, got[i].Status)
					}
				}
			}
		})
	}
}

// TestFindOrderByIDAndShopID - 注文IDと店舗IDで注文を取得するテスト
func TestFindOrderByIDAndShopID(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		orderID         int
		shopID          int
		setup           func(*sqlx.Tx)
		want            *models.Order
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:    "正常に注文を取得できる",
			orderID: testOrderID1,
			shopID:  testShopID1,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
			want: &models.Order{
				OrderID:     testOrderID1,
				UserID:      sql.NullInt64{Int64: int64(testUserID1), Valid: true},
				ShopID:      testShopID1,
				Status:      models.Cooking,
				TotalAmount: testTotalAmount1,
			},
		},
		{
			name:    "異なる店舗IDでは注文が取得できない",
			orderID: testOrderID1,
			shopID:  testShopID2,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
			want:            nil,
			expectedErrCode: apperrors.NoData,
		},
		{
			name:            "存在しない注文IDでは注文が取得できない",
			orderID:         nonExistentOrderID,
			shopID:          testShopID1,
			setup:           func(tx *sqlx.Tx) {},
			want:            nil,
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			got, err := repo.FindOrderByIDAndShopID(context.Background(), tt.orderID, tt.shopID)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				if tt.want != nil {
					// 注文の詳細を検証
					if got == nil {
						t.Error("got order is nil")
						return
					}
					if tt.want.OrderID != got.OrderID {
						t.Errorf("OrderID: expected %d, got %d", tt.want.OrderID, got.OrderID)
					}
					if tt.want.UserID != got.UserID {
						t.Errorf("UserID: expected %v, got %v", tt.want.UserID, got.UserID)
					}
					if tt.want.ShopID != got.ShopID {
						t.Errorf("ShopID: expected %d, got %d", tt.want.ShopID, got.ShopID)
					}
					if tt.want.Status != got.Status {
						t.Errorf("Status: expected %v, got %v", tt.want.Status, got.Status)
					}
					if tt.want.TotalAmount != got.TotalAmount {
						t.Errorf("TotalAmount: expected %f, got %f", tt.want.TotalAmount, got.TotalAmount)
					}
				} else {
					if got != nil {
						t.Errorf("expected nil order, got %+v", got)
					}
				}
			}
		})
	}
}

// TestUpdateOrderStatus - 注文ステータス更新テスト
func TestUpdateOrderStatus(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		orderID         int
		shopID          int
		newStatus       models.OrderStatus
		setup           func(*sqlx.Tx)
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:      "正常にステータスを更新できる",
			orderID:   testOrderID1,
			shopID:    testShopID1,
			newStatus: models.Cooking,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Completed)
			},
		},
		{
			name:      "異なる店舗IDでは更新できない",
			orderID:   testOrderID1,
			shopID:    testShopID2,
			newStatus: models.Cooking,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Completed)
			},
			expectedErrCode: apperrors.NoData,
		},
		{
			name:            "存在しない注文IDでは更新できない",
			orderID:         nonExistentOrderID,
			shopID:          testShopID1,
			newStatus:       models.Cooking,
			setup:           func(tx *sqlx.Tx) {},
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			err := repo.UpdateOrderStatus(context.Background(), tt.orderID, tt.shopID, tt.newStatus)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				// 正常に更新された場合は実際にステータスが変更されているか確認
				var status models.OrderStatus
				query := "SELECT status FROM orders WHERE order_id = $1 AND shop_id = $2"
				err := tx.Get(&status, query, tt.orderID, tt.shopID)
				if err != nil {
					t.Errorf("failed to verify status update: %v", err)
				} else if status != tt.newStatus {
					t.Errorf("expected status %v, got %v", tt.newStatus, status)
				}
			}
		})
	}
}

// TestDeleteOrderByIDAndShopID - 注文削除テスト
func TestDeleteOrderByIDAndShopID(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		orderID         int
		shopID          int
		setup           func(*sqlx.Tx)
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:    "正常に注文を削除できる",
			orderID: testOrderID1,
			shopID:  testShopID1,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
		},
		{
			name:    "異なる店舗IDでは削除できない",
			orderID: testOrderID1,
			shopID:  testShopID2,
			setup: func(tx *sqlx.Tx) {
				createTestUser(t, tx, testUserID1, fmt.Sprintf("user%d@test.com", testUserID1))
				createTestShop(t, tx, testShopID1, fmt.Sprintf("Test Shop %d", testShopID1))
				createTestOrder(t, tx, testOrderID1, testUserID1, testShopID1, models.Cooking)
			},
			expectedErrCode: apperrors.NoData,
		},
		{
			name:            "存在しない注文IDでは削除できない",
			orderID:         nonExistentOrderID,
			shopID:          testShopID1,
			setup:           func(tx *sqlx.Tx) {},
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewOrderRepository(tx)
			err := repo.DeleteOrderByIDAndShopID(context.Background(), tt.orderID, tt.shopID)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				// 正常に削除された場合は実際に注文が削除されているか確認
				var count int
				query := "SELECT COUNT(*) FROM orders WHERE order_id = $1 AND shop_id = $2"
				err = tx.Get(&count, query, tt.orderID, tt.shopID)
				if err != nil {
					t.Errorf("failed to verify deletion: %v", err)
				} else if count != 0 {
					t.Errorf("expected order to be deleted, but still exists")
				}
			}
		})
	}
}
