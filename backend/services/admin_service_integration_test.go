package services_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
=== AdminService DBTX対応 結合テスト ===

このファイルは、DBTX対応後のAdminServiceの結合テストを実装します。
主にトランザクション処理が含まれる機能をテストします。

【テスト対象】
- AdminService.UpdateOrderStatus() のトランザクション処理
- AdminService.DeleteOrder() のトランザクション処理
- データベース制約違反時のロールバック動作
- DBTX interfaceを使用したトランザクション境界
- 注文ステータス更新の一貫性

【単体テストとの分離】
- 単体テスト（admin_service_test.go）: ビジネスロジック、モック使用、高速実行
- 結合テスト（このファイル）: トランザクション、実DB使用、データ整合性
*/

// setupAdminTestDB はテスト用のデータベース接続をセットアップします
func setupAdminTestDB(t *testing.T) *sqlx.DB {
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// テスト実行時はlocalhost経由でアクセス
	dbURL = strings.ReplaceAll(dbURL, "@db:", "@localhost:")

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	return db
}

// createTestOrder テスト用注文を作成するヘルパー
func createTestOrder(t *testing.T, db *sqlx.DB, shopID int, status models.OrderStatus) int {
	// ユニークなゲストトークンを生成（タイムスタンプベース）
	guestToken := fmt.Sprintf("test-guest-token-%d", time.Now().UnixNano())

	var orderID int
	err := db.QueryRow(`
		INSERT INTO orders (shop_id, total_amount, status, guest_order_token, order_date) 
		VALUES ($1, 1000, $2, $3, NOW()) 
		RETURNING order_id
	`, shopID, status, guestToken).Scan(&orderID)
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}
	return orderID
}

// TestAdminService_UpdateOrderStatus_Integration 注文ステータス更新の結合テスト
func TestAdminService_UpdateOrderStatus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	db := setupAdminTestDB(t)
	defer db.Close()

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	adminService := services.NewAdminService(orderRepo, itemRepo, db)

	ctx := context.Background()

	t.Run("正常系: 調理中→調理完了へのステータス更新", func(t *testing.T) {
		// テストショップの存在確認
		var shopExists int
		err := db.QueryRow("SELECT COUNT(*) FROM shops WHERE shop_id = 1").Scan(&shopExists)
		if err != nil || shopExists == 0 {
			t.Skip("テストショップ(ID=1)が存在しないためスキップ")
		}

		// テスト用注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// ステータス更新実行
		err = adminService.UpdateOrderStatus(ctx, 1, orderID)
		testhelpers.AssertNoError(t, err)

		// DBでステータスが更新されていることを確認
		var actualStatus models.OrderStatus
		err = db.QueryRow("SELECT status FROM orders WHERE order_id = $1 AND shop_id = $2", orderID, 1).Scan(&actualStatus)
		if err != nil {
			t.Fatalf("Failed to get updated order status: %v", err)
		}

		if actualStatus != models.Completed {
			t.Errorf("Expected status=%v, got=%v", models.Completed, actualStatus)
		}
	})

	t.Run("正常系: 調理完了→受け渡し完了へのステータス更新", func(t *testing.T) {
		// テスト用注文を作成
		orderID := createTestOrder(t, db, 1, models.Completed)

		// ステータス更新実行
		err := adminService.UpdateOrderStatus(ctx, 1, orderID)
		testhelpers.AssertNoError(t, err)

		// DBでステータスが更新されていることを確認
		var actualStatus models.OrderStatus
		err = db.QueryRow("SELECT status FROM orders WHERE order_id = $1 AND shop_id = $2", orderID, 1).Scan(&actualStatus)
		if err != nil {
			t.Fatalf("Failed to get updated order status: %v", err)
		}

		if actualStatus != models.Handed {
			t.Errorf("Expected status=%v, got=%v", models.Handed, actualStatus)
		}
	})

	t.Run("異常系: 存在しない注文のステータス更新", func(t *testing.T) {
		// 存在しない注文IDでステータス更新を試行
		err := adminService.UpdateOrderStatus(ctx, 1, 99999)
		testhelpers.AssertAppError(t, err, apperrors.NoData)
	})

	t.Run("異常系: 別店舗の注文ステータス更新", func(t *testing.T) {
		// Shop1に注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// Shop2から注文ステータスの更新を試行
		err := adminService.UpdateOrderStatus(ctx, 2, orderID)
		testhelpers.AssertAppError(t, err, apperrors.NoData)

		// 元のステータスが変更されていないことを確認
		var actualStatus models.OrderStatus
		err = db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&actualStatus)
		if err != nil {
			t.Fatalf("Failed to get order status: %v", err)
		}

		if actualStatus != models.Cooking {
			t.Errorf("Status should not be changed: expected=%v, got=%v", models.Cooking, actualStatus)
		}
	})

	t.Run("異常系: 受け渡し完了済み注文のステータス更新", func(t *testing.T) {
		// 受け渡し完了済みの注文を作成
		orderID := createTestOrder(t, db, 1, models.Handed)

		// ステータス更新を試行（失敗するべき）
		err := adminService.UpdateOrderStatus(ctx, 1, orderID)
		testhelpers.AssertAppError(t, err, apperrors.Conflict)

		// ステータスが変更されていないことを確認
		var actualStatus models.OrderStatus
		err = db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&actualStatus)
		if err != nil {
			t.Fatalf("Failed to get order status: %v", err)
		}

		if actualStatus != models.Handed {
			t.Errorf("Status should not be changed: expected=%v, got=%v", models.Handed, actualStatus)
		}
	})
}

// TestAdminService_DeleteOrder_Integration 注文削除の結合テスト
func TestAdminService_DeleteOrder_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	db := setupAdminTestDB(t)
	defer db.Close()

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	adminService := services.NewAdminService(orderRepo, itemRepo, db)

	ctx := context.Background()

	t.Run("正常系: 注文の削除", func(t *testing.T) {
		// テスト用注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// 削除前に注文が存在することを確認
		var countBefore int
		err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", orderID).Scan(&countBefore)
		if err != nil {
			t.Fatalf("Failed to count orders: %v", err)
		}
		if countBefore != 1 {
			t.Fatalf("Expected 1 order before deletion, got %d", countBefore)
		}

		// 注文削除実行
		err = adminService.DeleteOrder(ctx, 1, orderID)
		testhelpers.AssertNoError(t, err)

		// 削除後に注文が存在しないことを確認
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", orderID).Scan(&countAfter)
		if err != nil {
			t.Fatalf("Failed to count orders: %v", err)
		}
		if countAfter != 0 {
			t.Errorf("Expected 0 orders after deletion, got %d", countAfter)
		}
	})

	t.Run("異常系: 存在しない注文の削除", func(t *testing.T) {
		// 存在しない注文IDで削除を試行
		err := adminService.DeleteOrder(ctx, 1, 99999)
		testhelpers.AssertAppError(t, err, apperrors.NoData)
	})

	t.Run("異常系: 別店舗の注文削除", func(t *testing.T) {
		// Shop1に注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// Shop2から注文削除を試行
		err := adminService.DeleteOrder(ctx, 2, orderID)
		testhelpers.AssertAppError(t, err, apperrors.NoData)

		// 注文が削除されていないことを確認
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", orderID).Scan(&countAfter)
		if err != nil {
			t.Fatalf("Failed to count orders: %v", err)
		}
		if countAfter != 1 {
			t.Errorf("Order should not be deleted: expected=1, got=%d", countAfter)
		}
	})
}

// TestAdminService_TransactionRollback_Integration トランザクションロールバックの結合テスト
func TestAdminService_TransactionRollback_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped in short mode")
	}

	db := setupAdminTestDB(t)
	defer db.Close()

	// リポジトリ初期化
	orderRepo := repositories.NewOrderRepository()
	itemRepo := repositories.NewItemRepository()

	// サービス初期化
	adminService := services.NewAdminService(orderRepo, itemRepo, db)

	ctx := context.Background()

	t.Run("異常系: UpdateOrderStatusでのトランザクションロールバック", func(t *testing.T) {
		// テスト用注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// 更新前のステータスを確認
		var statusBefore models.OrderStatus
		err := db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&statusBefore)
		if err != nil {
			t.Fatalf("Failed to get order status: %v", err)
		}

		// 無効なステータス更新を試行（存在しない注文ID）
		err = adminService.UpdateOrderStatus(ctx, 1, 99999)
		testhelpers.AssertAppError(t, err, apperrors.NoData)

		// 元の注文のステータスが変更されていないことを確認（ロールバック確認）
		var statusAfter models.OrderStatus
		err = db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&statusAfter)
		if err != nil {
			t.Fatalf("Failed to get order status: %v", err)
		}

		if statusBefore != statusAfter {
			t.Errorf("Status should not change on rollback: before=%v, after=%v", statusBefore, statusAfter)
		}
	})

	t.Run("異常系: DeleteOrderでのトランザクションロールバック", func(t *testing.T) {
		// テスト用注文を作成
		orderID := createTestOrder(t, db, 1, models.Cooking)

		// 削除前の注文数を確認
		var countBefore int
		err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", orderID).Scan(&countBefore)
		if err != nil {
			t.Fatalf("Failed to count orders: %v", err)
		}

		// 無効な削除を試行（存在しない注文ID）
		err = adminService.DeleteOrder(ctx, 1, 99999)
		testhelpers.AssertAppError(t, err, apperrors.NoData)

		// 元の注文が削除されていないことを確認（ロールバック確認）
		var countAfter int
		err = db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", orderID).Scan(&countAfter)
		if err != nil {
			t.Fatalf("Failed to count orders: %v", err)
		}

		if countBefore != countAfter {
			t.Errorf("Order count should not change on rollback: before=%d, after=%d", countBefore, countAfter)
		}
	})
}
