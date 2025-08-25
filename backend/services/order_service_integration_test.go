package services_test

import (
	"context"
	"os"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/internal/testhelpers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
=== OrderService DBTX対応 結合テスト ===

このファイルは、DBTX対応後のOrderServiceの結合テストを実装します。
主にトランザクション処理が含まれる機能をテストします。

【テスト対象】
- OrderService.CreateOrder() のトランザクション処理
- OrderService.CreateAuthenticatedOrder() のトランザクション処理
- データベース制約違反時のロールバック動作
- DBTX interfaceを使用したトランザクション境界
- マルチテーブル操作の一貫性

【単体テストとの分離】
- 単体テスト（order_service_test.go）: ビジネスロジック、モック使用、高速実行
- 結合テスト（このファイル）: トランザクション、実DB使用、データ整合性
*/

// setupOrderTestDB はテスト用のデータベース接続をセットアップします
func setupOrderTestDB(t *testing.T) *sqlx.DB {
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	return db
}

// ensureTestDataExists テストに必要な基本データが存在することを確認し、なければ作成します
func ensureTestDataExists(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// ユーザーの存在確認・作成
	var userCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = 1").Scan(&userCount)
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}

	if userCount == 0 {
		_, err = db.Exec(`
			INSERT INTO users (user_id, email, role) 
			VALUES (1, 'test@example.com', 1)
			ON CONFLICT (user_id) DO NOTHING
		`)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	// ショップの存在確認・作成
	var shopCount int
	err = db.QueryRow("SELECT COUNT(*) FROM shops WHERE shop_id = 1").Scan(&shopCount)
	if err != nil {
		t.Fatalf("Failed to check shop existence: %v", err)
	}

	if shopCount == 0 {
		_, err = db.Exec(`
			INSERT INTO shops (shop_id, name, description, location) 
			VALUES (1, 'テスト店舗', 'テスト用の店舗です', 'テスト住所')
			ON CONFLICT (shop_id) DO NOTHING
		`)
		if err != nil {
			t.Fatalf("Failed to create test shop: %v", err)
		}
	}

	// アイテムの存在確認・作成
	var itemCount int
	err = db.QueryRow("SELECT COUNT(*) FROM items WHERE item_id = 1").Scan(&itemCount)
	if err != nil {
		t.Fatalf("Failed to check item existence: %v", err)
	}

	if itemCount == 0 {
		_, err = db.Exec(`
			INSERT INTO items (item_id, item_name, description, price, is_available) 
			VALUES (1, 'テスト商品', 'テスト用の商品です', 500, true)
			ON CONFLICT (item_id) DO NOTHING
		`)
		if err != nil {
			t.Fatalf("Failed to create test item: %v", err)
		}
	}

	// ショップアイテムの関連付け確認・作成
	var shopItemCount int
	err = db.QueryRow("SELECT COUNT(*) FROM shop_item WHERE shop_id = 1 AND item_id = 1").Scan(&shopItemCount)
	if err != nil {
		t.Fatalf("Failed to check shop_item existence: %v", err)
	}

	if shopItemCount == 0 {
		_, err = db.Exec(`
			INSERT INTO shop_item (shop_id, item_id) 
			VALUES (1, 1)
			ON CONFLICT (shop_id, item_id) DO NOTHING
		`)
		if err != nil {
			t.Fatalf("Failed to create test shop_item: %v", err)
		}
	}
}

func TestOrderService_CreateOrder_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	// データベース接続
	db := setupOrderTestDB(t)
	defer db.Close()

	// テストデータの存在確認・作成
	ensureTestDataExists(t, db)

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	orderService := services.NewOrderService(orderRepo, itemRepo, db)

	ctx := context.Background()

	tests := []struct {
		name            string
		shopID          int
		items           []models.OrderItemRequest
		expectedErrCode apperrors.ErrCode
		validateOrder   func(t *testing.T, order *models.Order)
	}{
		{
			name:   "正常系: ゲスト注文作成",
			shopID: 1, // テストDBに存在することを前提
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1}, // テストDBに存在することを前提
			},
			expectedErrCode: "",
			validateOrder: func(t *testing.T, order *models.Order) {
				if order.OrderID <= 0 {
					t.Error("OrderID should be greater than 0")
				}
				if order.UserID.Valid {
					t.Error("Guest order should have null UserID")
				}
				if !order.GuestOrderToken.Valid || order.GuestOrderToken.String == "" {
					t.Error("Guest order should have valid GuestOrderToken")
				}
				if order.ShopID != 1 {
					t.Errorf("Expected ShopID=1, got %d", order.ShopID)
				}
				if order.Status != models.Cooking {
					t.Errorf("Expected Status=Cooking, got %v", order.Status)
				}
			},
		},
		{
			name:   "異常系: 存在しない店舗ID",
			shopID: 9999,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1},
			},
			expectedErrCode: apperrors.BadParam,
		},
		{
			name:   "異常系: 存在しない商品ID",
			shopID: 1,
			items: []models.OrderItemRequest{
				{ItemID: 9999, Quantity: 1},
			},
			expectedErrCode: apperrors.BadParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト実行
			order, err := orderService.CreateOrder(ctx, tt.shopID, tt.items)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				testhelpers.AssertNoError(t, err)
				if tt.validateOrder != nil {
					tt.validateOrder(t, order)
				}
			} else {
				testhelpers.AssertAppError(t, err, tt.expectedErrCode)
			}
		})
	}
}

func TestOrderService_CreateAuthenticatedOrder_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	// データベース接続
	db := setupOrderTestDB(t)
	defer db.Close()

	// テストデータの存在確認・作成
	ensureTestDataExists(t, db)

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	orderService := services.NewOrderService(orderRepo, itemRepo, db)

	ctx := context.Background()

	tests := []struct {
		name            string
		userID          int
		shopID          int
		items           []models.OrderItemRequest
		expectedErrCode apperrors.ErrCode
		validateOrder   func(t *testing.T, order *models.Order)
	}{
		{
			name:   "正常系: 認証済みユーザー注文作成",
			userID: 1, // テストDBに存在することを前提
			shopID: 1, // テストDBに存在することを前提
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1}, // テストDBに存在することを前提
			},
			expectedErrCode: "",
			validateOrder: func(t *testing.T, order *models.Order) {
				if order.OrderID <= 0 {
					t.Error("OrderID should be greater than 0")
				}
				if !order.UserID.Valid || order.UserID.Int64 != 1 {
					t.Errorf("Expected UserID=1, got %v", order.UserID)
				}
				if order.GuestOrderToken.Valid {
					t.Error("Authenticated order should have null GuestOrderToken")
				}
				if order.ShopID != 1 {
					t.Errorf("Expected ShopID=1, got %d", order.ShopID)
				}
				if order.Status != models.Cooking {
					t.Errorf("Expected Status=Cooking, got %v", order.Status)
				}
			},
		},
		{
			name:   "異常系: 存在しない店舗ID",
			userID: 1,
			shopID: 9999,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 1},
			},
			expectedErrCode: apperrors.BadParam,
		},
		{
			name:   "異常系: 存在しない商品ID",
			userID: 1,
			shopID: 1,
			items: []models.OrderItemRequest{
				{ItemID: 9999, Quantity: 1},
			},
			expectedErrCode: apperrors.BadParam,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト実行
			order, err := orderService.CreateAuthenticatedOrder(ctx, tt.userID, tt.shopID, tt.items)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				testhelpers.AssertNoError(t, err)
				if tt.validateOrder != nil {
					tt.validateOrder(t, order)
				}
			} else {
				testhelpers.AssertAppError(t, err, tt.expectedErrCode)
			}
		})
	}
}

// =============================
// トランザクション結合テスト群
// =============================

// TestOrderService_TransactionRollback_Integration トランザクションロールバックの結合テスト
func TestOrderService_TransactionRollback_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	db := setupOrderTestDB(t)
	defer db.Close()

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	orderService := services.NewOrderService(orderRepo, itemRepo, db)

	ctx := context.Background()

	t.Run("異常系: 商品検証失敗時のトランザクションロールバック", func(t *testing.T) {
		// 注文作成前の注文数を取得
		var beforeCount int
		err := db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&beforeCount)
		if err != nil {
			t.Fatalf("Failed to get order count: %v", err)
		}

		// 存在しない商品での注文作成を試行（トランザクション内で失敗するはず）
		_, err = orderService.CreateOrder(ctx, 1, []models.OrderItemRequest{
			{ItemID: 99999, Quantity: 1}, // 存在しない商品ID
		})

		// エラーが発生することを確認
		if err == nil {
			t.Fatal("Expected error but got nil")
		}

		// 注文作成後の注文数を取得
		var afterCount int
		err = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&afterCount)
		if err != nil {
			t.Fatalf("Failed to get order count: %v", err)
		}

		// ロールバックによって注文数が変わらないことを確認
		if beforeCount != afterCount {
			t.Errorf("Transaction rollback failed: beforeCount=%d, afterCount=%d", beforeCount, afterCount)
		}
	})
}

// TestOrderService_TransactionCommit_Integration トランザクションコミットの結合テスト
func TestOrderService_TransactionCommit_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	db := setupOrderTestDB(t)
	defer db.Close()

	// テストデータの存在確認・作成
	ensureTestDataExists(t, db)

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	orderService := services.NewOrderService(orderRepo, itemRepo, db)

	ctx := context.Background()

	t.Run("正常系: マルチテーブル操作のトランザクション一貫性確認", func(t *testing.T) {
		// 注文作成前の状態を取得
		var beforeOrderCount, beforeOrderItemCount int
		err := db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&beforeOrderCount)
		if err != nil {
			t.Fatalf("Failed to get order count: %v", err)
		}
		err = db.QueryRow("SELECT COUNT(*) FROM order_item").Scan(&beforeOrderItemCount)
		if err != nil {
			t.Fatalf("Failed to get order_item count: %v", err)
		}

		// 注文作成
		items := []models.OrderItemRequest{
			{ItemID: 1, Quantity: 2},
		}

		order, err := orderService.CreateOrder(ctx, 1, items)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}

		// トランザクション一貫性の確認
		var afterOrderCount, afterOrderItemCount int
		err = db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&afterOrderCount)
		if err != nil {
			t.Fatalf("Failed to get order count: %v", err)
		}
		err = db.QueryRow("SELECT COUNT(*) FROM order_item WHERE order_id = $1", order.OrderID).Scan(&afterOrderItemCount)
		if err != nil {
			t.Fatalf("Failed to get order_item count: %v", err)
		}

		// 結果検証
		if afterOrderCount != beforeOrderCount+1 {
			t.Errorf("Orders table: expected count=%d, got=%d", beforeOrderCount+1, afterOrderCount)
		}
		if afterOrderItemCount != len(items) {
			t.Errorf("Order_item table: expected count=%d, got=%d", len(items), afterOrderItemCount)
		}
	})
}
