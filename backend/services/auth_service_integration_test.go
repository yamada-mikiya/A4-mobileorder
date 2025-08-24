package services_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
=== AuthService DBTX対応 結合テスト ===

このファイルは、DBTX対応後のAuthServiceの結合テストを実装します。
主にトランザクション処理が含まれる機能をテストします。

【テスト対象】
- AuthService.SignUp() のトランザクション処理（ゲストトークンあり）
- AuthService.LogIn() のトランザクション処理（ゲストトークンあり）
- データベース制約違反時のロールバック動作
- DBTX interfaceを使用したトランザクション境界

【単体テストとの分離】
- 単体テスト（auth_service_test.go）: ビジネスロジック、モック使用、高速実行
- 結合テスト（このファイル）: トランザクション、実DB使用、データ整合性
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

	// テスト実行時はlocalhost経由でアクセス
	dbURL = strings.ReplaceAll(dbURL, "@db:", "@localhost:")

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

	// 管理者ユーザー用のテストデータ
	_, err = db.Exec(`
		INSERT INTO users (user_id, email, role, created_at, updated_at)
		VALUES (999, 'admin@test.com', 2, NOW(), NOW())
		ON CONFLICT (user_id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("テスト管理者ユーザー作成失敗: %v", err)
	}

	// 管理者とショップの関連
	_, err = db.Exec(`
		INSERT INTO shop_admins (shop_id, admin_user_id, created_at, updated_at)
		VALUES (1, 999, NOW(), NOW())
		ON CONFLICT (shop_id, admin_user_id) DO NOTHING
	`)
	if err != nil {
		t.Logf("Warning: テスト管理者関連作成失敗: %v", err)
	}
}

// TestAuthService_SignUp_Integration AuthService.SignUp()のDBTX対応結合テスト
func TestAuthService_SignUp_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のリポジトリとサービスを作成（DBTX対応）
	userRepo := repositories.NewUserRepository()
	shopRepo := repositories.NewShopRepository()
	orderRepo := repositories.NewOrderRepository()
	authService := services.NewAuthService(userRepo, shopRepo, orderRepo, db)

	tests := []struct {
		name             string
		setupData        func() models.AuthenticateRequest
		validate         func(t *testing.T, userResponse models.UserResponse, token string, req models.AuthenticateRequest)
		validateRollback func(t *testing.T, req models.AuthenticateRequest) // ロールバックの検証
		expectError      bool
	}{
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンなし）",
			setupData: func() models.AuthenticateRequest {
				return models.AuthenticateRequest{Email: "newuser@example.com"}
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string, req models.AuthenticateRequest) {
				if userResponse.UserID == 0 {
					t.Error("UserIDが設定されていません")
				}
				if token == "" {
					t.Error("Tokenが生成されていません")
				}
				if userResponse.Role != models.CustomerRole.String() {
					t.Errorf("期待されるロール: %s, 実際: %s", models.CustomerRole.String(), userResponse.Role)
				}

				// DBでユーザーが作成されていることを確認
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
				if err != nil || count != 1 {
					t.Errorf("ユーザーがDBに作成されていません: count=%d, err=%v", count, err)
				}
			},
			expectError: false,
		},
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンあり - DBTX トランザクション処理）",
			setupData: func() models.AuthenticateRequest {
				// ゲスト注文を作成
				var orderID int
				err := db.QueryRow(`
					INSERT INTO orders (shop_id, order_date, total_amount, status, guest_order_token, created_at, updated_at)
					VALUES (1, NOW(), 500, 1, 'test-guest-token-123', NOW(), NOW())
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
			validate: func(t *testing.T, userResponse models.UserResponse, token string, req models.AuthenticateRequest) {
				if userResponse.UserID == 0 {
					t.Error("UserIDが設定されていません")
				}
				if token == "" {
					t.Error("Tokenが生成されていません")
				}

				// 【重要】ゲスト注文がユーザーにリンクされていることを確認（DBTX トランザクション成功の証明）
				var userID sql.NullInt64
				err := db.QueryRow("SELECT user_id FROM orders WHERE guest_order_token = $1", req.GuestOrderToken).Scan(&userID)
				if err != nil || !userID.Valid || userID.Int64 != int64(userResponse.UserID) {
					t.Errorf("DBTX トランザクション失敗: ゲスト注文がユーザーにリンクされていません: userID=%v, responseUserID=%d", userID, userResponse.UserID)
				}

				// ユーザーも正しく作成されていることを確認
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1 AND user_id = $2", req.Email, userResponse.UserID).Scan(&count)
				if err != nil || count != 1 {
					t.Errorf("ユーザーがDBに作成されていません: count=%d, err=%v", count, err)
				}
			},
			expectError: false,
		},
		{
			name: "異常系: 存在しないゲストトークン（DBTX トランザクションロールバック）",
			setupData: func() models.AuthenticateRequest {
				// 存在しないゲストトークンを使用
				return models.AuthenticateRequest{
					Email:           "rollbacktest@example.com",
					GuestOrderToken: "non-existent-token-456",
				}
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string, req models.AuthenticateRequest) {
				// このケースは呼ばれない（エラーのため）
			},
			validateRollback: func(t *testing.T, req models.AuthenticateRequest) {
				// 【重要】DBTX ロールバックが正しく動作していることを確認
				var count int
				dbErr := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
				if dbErr != nil {
					t.Errorf("ロールバック検証のクエリ実行エラー: %v", dbErr)
				}
				if count != 0 {
					t.Errorf("DBTX ロールバックが正しく動作していません。ユーザーが残っています: count=%d", count)
				}
			},
			expectError: true,
		},
		{
			name: "異常系: 重複メールアドレス（DBTX トランザクションロールバック）",
			setupData: func() models.AuthenticateRequest {
				// 重複するメールのユーザーを事前作成
				_, err := db.Exec(`
					INSERT INTO users (email, created_at, updated_at)
					VALUES ('duplicate@example.com', NOW(), NOW())
				`)
				if err != nil {
					t.Fatalf("重複テスト用ユーザー作成失敗: %v", err)
				}

				// ゲスト注文も作成（ロールバック対象）
				_, err = db.Exec(`
					INSERT INTO orders (shop_id, order_date, total_amount, status, guest_order_token, created_at, updated_at)
					VALUES (1, NOW(), 300, 1, 'duplicate-test-token', NOW(), NOW())
				`)
				if err != nil {
					t.Fatalf("重複テスト用ゲスト注文作成失敗: %v", err)
				}

				return models.AuthenticateRequest{
					Email:           "duplicate@example.com",
					GuestOrderToken: "duplicate-test-token",
				}
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string, req models.AuthenticateRequest) {
				// このケースは呼ばれない（エラーのため）
			},
			validateRollback: func(t *testing.T, req models.AuthenticateRequest) {
				// 【重要】重複エラーでもDBTX ロールバックが正しく動作していることを確認
				var count int
				dbErr := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
				if dbErr != nil {
					t.Errorf("ロールバック検証のクエリ実行エラー: %v", dbErr)
				}
				if count != 1 {
					t.Errorf("重複メールのユーザー数が正しくありません。期待値: 1, 実際: %d", count)
				}

				// ゲスト注文のuser_idがnullのままであることを確認（リンクが実行されていない）
				var userID sql.NullInt64
				dbErr = db.QueryRow("SELECT user_id FROM orders WHERE guest_order_token = $1", req.GuestOrderToken).Scan(&userID)
				if dbErr != nil {
					t.Errorf("ゲスト注文確認のクエリ実行エラー: %v", dbErr)
				}
				if userID.Valid {
					t.Errorf("DBTX ロールバック失敗: ゲスト注文がリンクされています: userID=%v", userID)
				}
			},
			expectError: true,
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
				// ロールバック検証
				if tt.validateRollback != nil {
					tt.validateRollback(t, req)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				} else {
					tt.validate(t, userResponse, token, req)
				}
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

// TestAuthService_LogIn_Integration AuthService.LogIn()のDBTX対応結合テスト
func TestAuthService_LogIn_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test skipped with -short flag")
	}

	db := setupTestDB(t)
	defer db.Close()

	// 実際のリポジトリとサービスを作成（DBTX対応）
	userRepo := repositories.NewUserRepository()
	shopRepo := repositories.NewShopRepository()
	orderRepo := repositories.NewOrderRepository()
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
		VALUES (1, NOW(), 800, 1, 'login-guest-token-456', NOW(), NOW())
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
			name: "正常系: 既存ユーザーログイン（ゲストトークンなし）",
			req: models.AuthenticateRequest{
				Email: "existing@example.com",
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string) {
				if userResponse.UserID != userID {
					t.Errorf("UserIDが正しくありません: expected=%d, actual=%d", userID, userResponse.UserID)
				}
				if token == "" {
					t.Error("Tokenが生成されていません")
				}
				if userResponse.Role != models.CustomerRole.String() {
					t.Errorf("期待されるロール: %s, 実際: %s", models.CustomerRole.String(), userResponse.Role)
				}
			},
			expectError: false,
		},
		{
			name: "正常系: 既存ユーザーログイン（ゲストトークンあり - DBTX トランザクション処理）",
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

				// 【重要】ゲスト注文がユーザーにリンクされていることを確認（DBTX トランザクション成功の証明）
				var linkedUserID sql.NullInt64
				err := db.QueryRow("SELECT user_id FROM orders WHERE order_id = $1", guestOrderID).Scan(&linkedUserID)
				if err != nil || !linkedUserID.Valid || linkedUserID.Int64 != int64(userID) {
					t.Errorf("DBTX トランザクション失敗: ゲスト注文がユーザーにリンクされていません: userID=%v, expectedUserID=%d", linkedUserID, userID)
				}
			},
			expectError: false,
		},
		{
			name: "異常系: 存在しないユーザー",
			req: models.AuthenticateRequest{
				Email: "nonexistent@example.com",
			},
			validate: func(t *testing.T, userResponse models.UserResponse, token string) {
				// このケースは呼ばれない（エラーのため）
			},
			expectError: true,
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
				} else {
					tt.validate(t, userResponse, token)
				}
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
