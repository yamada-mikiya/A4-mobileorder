package services_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
=== サービス層統合テスト対象関数 ===

【インフラストラクチャレベルテスト】
- TestTransactionManager_WithFullTransaction_Success: 複数リポジトリを使った正常なトランザクションのコミット
- TestTransactionManager_WithFullTransaction_Rollback: 複数リポジトリでエラー発生時のロールバック
- TestTransactionManager_WithFullTransaction_UserCreateFails: ユーザー作成失敗時のロールバック
- TestTransactionManager_WithTransaction_SingleRepository: 単一リポジトリでの正常なトランザクション
- TestTransactionManager_WithTransaction_Rollback: 単一リポジトリでエラー発生時のロールバック

【サービスレベル統合テスト - 実装済み】
AuthService.SignUp() - TestAuthService_SignUp_Integration
- ユーザー作成とゲスト注文リンクのトランザクション動作

AuthService.LogIn() - TestAuthService_LogIn_Integration
- ユーザーログインとゲスト注文リンクのトランザクション動作

OrderService.CreateOrder() - TestOrderService_CreateOrder_Integration
- ゲスト注文作成のトランザクション動作

AdminService.UpdateOrderStatus() - TestAdminService_UpdateOrderStatus_Integration
- 注文ステータス更新のトランザクション動作
*/

// setupTestDB はテスト用のデータベース接続をセットアップします
func setupTestDB(t *testing.T) *sqlx.DB {
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// テスト用スキーマのロード
	schema, err := os.ReadFile("../repositories/testdata/schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// テスト用基本データの作成
	setupTestData(t, db)

	return db
}

// setupTestData テスト用の基本データを作成
func setupTestData(t *testing.T, db *sqlx.DB) {
	// ショップデータを作成
	_, err := db.Exec(`
		INSERT INTO shops (shop_id, name, description, location, is_open, created_at, updated_at)
		VALUES (1, 'テストショップ', 'テスト用ショップ', 'テスト場所', true, NOW(), NOW())
		ON CONFLICT (shop_id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("テストショップ作成失敗: %v", err)
	}

	// 商品データを作成
	_, err = db.Exec(`
		INSERT INTO items (item_id, item_name, description, price, is_available, created_at, updated_at)
		VALUES 
			(1, 'テスト商品1', 'テスト用商品1', 100.0, true, NOW(), NOW()),
			(2, 'テスト商品2', 'テスト用商品2', 200.0, true, NOW(), NOW())
		ON CONFLICT (item_id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("テスト商品作成失敗: %v", err)
	}

	// ショップ商品関連を作成
	_, err = db.Exec(`
		INSERT INTO shop_item (shop_id, item_id)
		VALUES (1, 1), (1, 2)
		ON CONFLICT (shop_id, item_id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("ショップ商品関連作成失敗: %v", err)
	}
}

// TestTransactionManager_WithFullTransaction_Success 複数リポジトリを使った正常なトランザクションのテスト
func TestTransactionManager_WithFullTransaction_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	txManager := services.NewTransactionManager(db)

	// テスト用のユーザーとゲスト注文を事前作成
	var guestOrderID int
	err := db.QueryRow(`
		INSERT INTO orders (shop_id, order_date, total_amount, status, guest_order_token, created_at, updated_at)
		VALUES (1, NOW(), 500.0, 1, 'test-guest-token', NOW(), NOW())
		RETURNING order_id
	`).Scan(&guestOrderID)
	if err != nil {
		t.Fatalf("ゲスト注文作成失敗: %v", err)
	}

	var userID int
	err = txManager.WithFullTransaction(context.Background(), func(userRepo repositories.UserRepository, orderRepo repositories.OrderRepository) error {
		// 1. ユーザーを作成
		user := &models.User{Email: "test@example.com"}
		if err := userRepo.CreateUser(context.Background(), user); err != nil {
			return err
		}
		userID = user.UserID

		// 2. ゲスト注文をユーザーにリンク
		if err := orderRepo.UpdateUserIDByGuestToken(context.Background(), "test-guest-token", userID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("トランザクション実行失敗: %v", err)
	}

	// 結果検証: ユーザーが作成されていること
	var createdEmail string
	err = db.QueryRow("SELECT email FROM users WHERE user_id = $1", userID).Scan(&createdEmail)
	if err != nil {
		t.Fatalf("ユーザー検証失敗: %v", err)
	}
	if createdEmail != "test@example.com" {
		t.Errorf("期待されるメール: test@example.com, 実際: %s", createdEmail)
	}

	// 結果検証: ゲスト注文がリンクされていること
	var linkedUserID sql.NullInt64
	err = db.QueryRow("SELECT user_id FROM orders WHERE order_id = $1", guestOrderID).Scan(&linkedUserID)
	if err != nil {
		t.Fatalf("注文リンク検証失敗: %v", err)
	}
	if !linkedUserID.Valid || linkedUserID.Int64 != int64(userID) {
		t.Errorf("ゲスト注文がユーザーにリンクされていません: userID=%v, expectedUserID=%d", linkedUserID, userID)
	}

	// クリーンアップ
	if _, err := db.Exec("DELETE FROM orders WHERE order_id = $1", guestOrderID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users WHERE user_id = $1", userID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
}

// TestTransactionManager_WithFullTransaction_Rollback エラー発生時のロールバックテスト
func TestTransactionManager_WithFullTransaction_Rollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	txManager := services.NewTransactionManager(db)

	// 存在しないゲストトークンでエラーを意図的に発生させる
	err := txManager.WithFullTransaction(context.Background(), func(userRepo repositories.UserRepository, orderRepo repositories.OrderRepository) error {
		// 1. ユーザーを作成
		user := &models.User{Email: "rollback-test@example.com"}
		if err := userRepo.CreateUser(context.Background(), user); err != nil {
			return err
		}

		// 2. 存在しないゲストトークンでエラーを発生させる
		if err := orderRepo.UpdateUserIDByGuestToken(context.Background(), "non-existent-token", user.UserID); err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		t.Error("エラーが期待されましたが、nilが返されました")
	}
	t.Logf("Transaction error (expected): %v", err)

	// ロールバック検証: ユーザーが作成されていないこと
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "rollback-test@example.com").Scan(&count); err != nil {
		t.Errorf("ロールバック検証のクエリ実行エラー: %v", err)
	}
	if count != 0 {
		t.Errorf("ロールバックが正しく動作していません。ユーザーが残っています: count=%d", count)
	}
}

// TestTransactionManager_WithFullTransaction_UserCreateFails ユーザー作成失敗時のロールバックテスト
func TestTransactionManager_WithFullTransaction_UserCreateFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	txManager := services.NewTransactionManager(db)

	// 同じメールアドレスのユーザーを事前作成
	var existingUserID int
	err := db.QueryRow(`
		INSERT INTO users (email, created_at, updated_at)
		VALUES ('duplicate@example.com', NOW(), NOW())
		RETURNING user_id
	`).Scan(&existingUserID)
	if err != nil {
		t.Fatalf("事前ユーザー作成失敗: %v", err)
	}

	// 重複メールでエラーを発生させる
	err = txManager.WithFullTransaction(context.Background(), func(userRepo repositories.UserRepository, orderRepo repositories.OrderRepository) error {
		// 重複メールでユーザー作成を試行
		user := &models.User{Email: "duplicate@example.com"}
		return userRepo.CreateUser(context.Background(), user)
	})

	if err == nil {
		t.Error("エラーが期待されましたが、nilが返されました")
	}

	// 重複したユーザーは1件のみであることを確認
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "duplicate@example.com").Scan(&count); err != nil {
		t.Errorf("ユーザー数確認のクエリ実行エラー: %v", err)
	}
	if count != 1 {
		t.Errorf("ユーザー数が正しくありません。期待値: 1, 実際: %d", count)
	}

	// クリーンアップ
	if _, err := db.Exec("DELETE FROM users WHERE user_id = $1", existingUserID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
}

// TestTransactionManager_WithTransaction_SingleRepository 単一リポジトリでのトランザクションテスト
func TestTransactionManager_WithTransaction_SingleRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	txManager := services.NewTransactionManager(db)

	var orderID int
	err := txManager.WithTransaction(context.Background(), func(orderRepo repositories.OrderRepository) error {
		// テスト用注文を作成
		order := &models.Order{
			ShopID:          1,
			TotalAmount:     300.0,
			Status:          models.Cooking,
			GuestOrderToken: sql.NullString{String: "single-repo-token", Valid: true},
		}

		orderItems := []models.OrderItem{
			{ItemID: 1, Quantity: 2, PriceAtOrder: 100.0},
			{ItemID: 2, Quantity: 1, PriceAtOrder: 100.0},
		}

		if err := orderRepo.CreateOrder(context.Background(), order, orderItems); err != nil {
			return err
		}
		orderID = order.OrderID
		return nil
	})

	if err != nil {
		t.Fatalf("トランザクション実行失敗: %v", err)
	}

	// 結果検証: 注文が作成されていること
	var totalAmount float64
	err = db.QueryRow("SELECT total_amount FROM orders WHERE order_id = $1", orderID).Scan(&totalAmount)
	if err != nil {
		t.Fatalf("注文検証失敗: %v", err)
	}
	if totalAmount != 300.0 {
		t.Errorf("期待される合計額: 300.0, 実際: %.2f", totalAmount)
	}

	// クリーンアップ
	if _, err := db.Exec("DELETE FROM order_item WHERE order_id = $1", orderID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
	if _, err := db.Exec("DELETE FROM orders WHERE order_id = $1", orderID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
}

// TestTransactionManager_WithTransaction_Rollback 単一リポジトリでのロールバックテスト
func TestTransactionManager_WithTransaction_Rollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	txManager := services.NewTransactionManager(db)

	// 意図的にエラーを発生させる
	err := txManager.WithTransaction(context.Background(), func(orderRepo repositories.OrderRepository) error {
		// 正常な注文作成
		order := &models.Order{
			ShopID:          1,
			TotalAmount:     400.0,
			Status:          models.Cooking,
			GuestOrderToken: sql.NullString{String: "rollback-token", Valid: true},
		}

		orderItems := []models.OrderItem{
			{ItemID: 1, Quantity: 3, PriceAtOrder: 100.0},
		}

		if err := orderRepo.CreateOrder(context.Background(), order, orderItems); err != nil {
			return err
		}

		// 意図的にエラーを発生させてロールバックをテスト
		return errors.New("意図的なエラー")
	})

	if err == nil {
		t.Error("エラーが期待されましたが、nilが返されました")
	}

	// ロールバック検証: 注文が作成されていないこと
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE guest_order_token = $1", "rollback-token").Scan(&count); err != nil {
		t.Errorf("ロールバック検証のクエリ実行エラー: %v", err)
	}
	if count != 0 {
		t.Errorf("ロールバックが正しく動作していません。注文が残っています: count=%d", count)
	}
}

// TestAuthService_SignUp_Integration AuthService.SignUp()の統合テスト
func TestAuthService_SignUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のリポジトリとサービスを作成
	userRepo := repositories.NewUserRepository(db)
	shopRepo := repositories.NewShopRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	authService := services.NewAuthService(userRepo, shopRepo, orderRepo, db)

	tests := []struct {
		name        string
		req         models.AuthenticateRequest
		setupData   func() models.AuthenticateRequest // リクエストセットアップ
		validate    func(t *testing.T, userResponse models.UserResponse, token string)
		expectError bool
	}{
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンなし）",
			req:  models.AuthenticateRequest{Email: "newuser@example.com"},
			setupData: func() models.AuthenticateRequest {
				return models.AuthenticateRequest{Email: "newuser@example.com"}
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string) {
				if userResponse.UserID == 0 {
					t.Error("UserIDが設定されていません")
				}
				if token == "" {
					t.Error("Tokenが生成されていません")
				}
				// DBでユーザーが作成されていることを確認
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "newuser@example.com").Scan(&count)
				if err != nil || count != 1 {
					t.Errorf("ユーザーがDBに作成されていません: count=%d, err=%v", count, err)
				}
			},
			expectError: false,
		},
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンあり）",
			req:  models.AuthenticateRequest{Email: "userWithGuest@example.com"},
			setupData: func() models.AuthenticateRequest {
				// ゲスト注文を作成
				var orderID int
				err := db.QueryRow(`
					INSERT INTO orders (shop_id, order_date, total_amount, status, guest_order_token, created_at, updated_at)
					VALUES (1, NOW(), 500.0, 1, 'test-guest-token-123', NOW(), NOW())
					RETURNING order_id
				`).Scan(&orderID)
				if err != nil {
					t.Fatalf("ゲスト注文作成失敗: %v", err)
				}
				return models.AuthenticateRequest{
					Email:           "userWithGuest@example.com",
					GuestOrderToken: "test-guest-token-123",
				}
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string) {
				if userResponse.UserID == 0 {
					t.Error("UserIDが設定されていません")
				}

				// ゲスト注文がユーザーにリンクされていることを確認
				var userID sql.NullInt64
				err := db.QueryRow("SELECT user_id FROM orders WHERE guest_order_token = $1", "test-guest-token-123").Scan(&userID)
				if err != nil || !userID.Valid || userID.Int64 != int64(userResponse.UserID) {
					t.Errorf("ゲスト注文がユーザーにリンクされていません: userID=%v, responseUserID=%d", userID, userResponse.UserID)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// データセットアップ
			req := tt.setupData()

			// テスト実行
			userResponse, token, err := authService.SignUp(context.Background(), req)

			// エラーアサーション
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				tt.validate(t, userResponse, token)
			}

			// テストデータクリーンアップ
			if _, err := db.Exec("DELETE FROM orders WHERE guest_order_token = $1", req.GuestOrderToken); err != nil {
				t.Logf("クリーンアップエラー: %v", err)
			}
			if _, err := db.Exec("DELETE FROM users WHERE email = $1", req.Email); err != nil {
				t.Logf("クリーンアップエラー: %v", err)
			}
		})
	}
}

// TestOrderService_CreateOrder_Integration OrderService.CreateOrder()の統合テスト
func TestOrderService_CreateOrder_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のリポジトリとサービスを作成
	orderRepo := repositories.NewOrderRepository(db)
	itemRepo := repositories.NewItemRepository(db)
	orderService := services.NewOrderService(orderRepo, itemRepo, db)

	tests := []struct {
		name        string
		shopID      int
		items       []models.OrderItemRequest
		validate    func(t *testing.T, order *models.Order)
		expectError bool
	}{
		{
			name:   "正常系: ゲスト注文作成とトランザクション",
			shopID: 1,
			items: []models.OrderItemRequest{
				{ItemID: 1, Quantity: 2}, // テスト商品1: 100円 × 2
				{ItemID: 2, Quantity: 1}, // テスト商品2: 200円 × 1
			},
			validate: func(t *testing.T, order *models.Order) {
				expectedTotal := 100.0*2 + 200.0*1 // 400.0
				if order.TotalAmount != expectedTotal {
					t.Errorf("合計金額が正しくありません: expected=%.2f, actual=%.2f", expectedTotal, order.TotalAmount)
				}
				if !order.GuestOrderToken.Valid || order.GuestOrderToken.String == "" {
					t.Error("ゲストトークンが設定されていません")
				}
				if order.UserID.Valid {
					t.Error("ゲスト注文ではuserIDはnullであるべきです")
				}

				// DBで注文とアイテムが作成されていることを確認
				var orderCount, itemCount int
				err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE order_id = $1", order.OrderID).Scan(&orderCount)
				if err != nil {
					t.Errorf("注文カウント取得エラー: %v", err)
				}
				err = db.QueryRow("SELECT COUNT(*) FROM order_item WHERE order_id = $1", order.OrderID).Scan(&itemCount)
				if err != nil {
					t.Errorf("注文アイテムカウント取得エラー: %v", err)
				}

				if orderCount != 1 {
					t.Errorf("注文がDBに作成されていません: count=%d", orderCount)
				}
				if itemCount != 2 {
					t.Errorf("注文アイテムが正しく作成されていません: count=%d", itemCount)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト実行
			order, err := orderService.CreateOrder(context.Background(), tt.shopID, tt.items)

			// エラーアサーション
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				tt.validate(t, order)
			}

			// クリーンアップ
			if order != nil {
				if _, err := db.Exec("DELETE FROM order_item WHERE order_id = $1", order.OrderID); err != nil {
					t.Logf("クリーンアップエラー: %v", err)
				}
				if _, err := db.Exec("DELETE FROM orders WHERE order_id = $1", order.OrderID); err != nil {
					t.Logf("クリーンアップエラー: %v", err)
				}
			}
		})
	}
}

// TestAdminService_UpdateOrderStatus_Integration AdminService.UpdateOrderStatus()の統合テスト
func TestAdminService_UpdateOrderStatus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のサービスを作成
	adminService := services.NewAdminService(db)

	tests := []struct {
		name        string
		orderID     int
		shopID      int
		setupData   func() int // 注文IDを返す
		validate    func(t *testing.T, orderID int)
		expectError bool
	}{
		{
			name:    "正常系: 注文ステータス更新",
			orderID: 0, // setupDataで設定
			shopID:  1,
			setupData: func() int {
				// テスト注文を作成
				var orderID int
				err := db.QueryRow(`
					INSERT INTO orders (shop_id, order_date, total_amount, status, created_at, updated_at)
					VALUES (1, NOW(), 1000.0, 1, NOW(), NOW())
					RETURNING order_id
				`).Scan(&orderID)
				if err != nil {
					t.Fatalf("テスト注文作成失敗: %v", err)
				}
				return orderID
			},
			validate: func(t *testing.T, orderID int) {
				// ステータスが進歩していることを確認
				var status int
				err := db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&status)
				if err != nil {
					t.Errorf("注文ステータス取得失敗: %v", err)
				}
				if status != 2 { // 1から2に進歩
					t.Errorf("期待されるステータス: 2, 実際: %d", status)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// データセットアップ
			orderID := tt.setupData()

			// テスト実行
			err := adminService.UpdateOrderStatus(context.Background(), tt.shopID, orderID)

			// エラーアサーション
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				tt.validate(t, orderID)
			}

			// テストデータクリーンアップ
			if _, err := db.Exec("DELETE FROM orders WHERE order_id = $1", orderID); err != nil {
				t.Logf("クリーンアップエラー: %v", err)
			}
		})
	}
}

// TestAuthService_LogIn_Integration AuthService.LogIn()の統合テスト
func TestAuthService_LogIn_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のリポジトリとサービスを作成
	userRepo := repositories.NewUserRepository(db)
	shopRepo := repositories.NewShopRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	authService := services.NewAuthService(userRepo, shopRepo, orderRepo, db)

	// テスト用ユーザーとゲスト注文を事前作成
	var userID int
	err := db.QueryRow(`
		INSERT INTO users (email, created_at, updated_at)
		VALUES ('existing@example.com', NOW(), NOW())
		RETURNING user_id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	var guestOrderID int
	err = db.QueryRow(`
		INSERT INTO orders (shop_id, order_date, total_amount, status, guest_order_token, created_at, updated_at)
		VALUES (1, NOW(), 800.0, 1, 'login-guest-token-456', NOW(), NOW())
		RETURNING order_id
	`).Scan(&guestOrderID)
	if err != nil {
		t.Fatalf("ゲスト注文作成失敗: %v", err)
	}

	tests := []struct {
		name        string
		req         models.AuthenticateRequest
		validate    func(t *testing.T, userResponse models.UserResponse, token string)
		expectError bool
	}{
		{
			name: "正常系: 既存ユーザーログイン（ゲストトークンあり）",
			req: models.AuthenticateRequest{
				Email:           "existing@example.com",
				GuestOrderToken: "login-guest-token-456",
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string) {
				if userResponse.UserID != userID {
					t.Errorf("UserIDが正しくありません: expected=%d, actual=%d", userID, userResponse.UserID)
				}
				if token == "" {
					t.Error("Tokenが生成されていません")
				}

				// ゲスト注文がユーザーにリンクされていることを確認
				var linkedUserID sql.NullInt64
				err := db.QueryRow("SELECT user_id FROM orders WHERE order_id = $1", guestOrderID).Scan(&linkedUserID)
				if err != nil || !linkedUserID.Valid || linkedUserID.Int64 != int64(userID) {
					t.Errorf("ゲスト注文がユーザーにリンクされていません: userID=%v, expectedUserID=%d", linkedUserID, userID)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト実行
			userResponse, token, err := authService.LogIn(context.Background(), tt.req)

			// エラーアサーション
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				tt.validate(t, userResponse, token)
			}
		})
	}

	// クリーンアップ
	if _, err := db.Exec("DELETE FROM orders WHERE order_id = $1", guestOrderID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
	if _, err := db.Exec("DELETE FROM users WHERE user_id = $1", userID); err != nil {
		t.Logf("クリーンアップエラー: %v", err)
	}
}
