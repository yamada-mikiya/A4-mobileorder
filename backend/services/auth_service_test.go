package services_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/internal/testhelpers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// UserRepositoryMockForAuth - UserRepositoryのモック実装（Auth用、DBTX対応）
type UserRepositoryMockForAuth struct {
	CreateUserFunc     func(ctx context.Context, dbtx repositories.DBTX, user *models.User) error
	GetUserByEmailFunc func(ctx context.Context, dbtx repositories.DBTX, email string) (models.User, error)
}

// NewUserRepositoryMockForAuth モック実装を返す
func NewUserRepositoryMockForAuth() *UserRepositoryMockForAuth {
	return &UserRepositoryMockForAuth{}
}

// インターフェース実装
func (m *UserRepositoryMockForAuth) CreateUser(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, dbtx, user)
	}
	panic("not implemented")
}

func (m *UserRepositoryMockForAuth) GetUserByEmail(ctx context.Context, dbtx repositories.DBTX, email string) (models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, dbtx, email)
	}
	panic("not implemented")
}

func (m *UserRepositoryMockForAuth) FindUserByID(ctx context.Context, dbtx repositories.DBTX, userID int) (*models.User, error) {
	panic("not implemented")
}

func (m *UserRepositoryMockForAuth) UpdateUser(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
	panic("not implemented")
}

func (m *UserRepositoryMockForAuth) DeleteUser(ctx context.Context, dbtx repositories.DBTX, userID int) error {
	panic("not implemented")
}

// ShopRepositoryMockForAuth - ShopRepositoryのモック実装（Auth用、DBTX対応）
type ShopRepositoryMockForAuth struct {
	FindShopIDByAdminIDFunc func(ctx context.Context, dbtx repositories.DBTX, adminID int) (int, error)
}

// NewShopRepositoryMockForAuth モック実装を返す
func NewShopRepositoryMockForAuth() *ShopRepositoryMockForAuth {
	return &ShopRepositoryMockForAuth{}
}

// インターフェース実装
func (m *ShopRepositoryMockForAuth) CreateShop(ctx context.Context, dbtx repositories.DBTX, shop *models.Shop) error {
	panic("not implemented")
}

func (m *ShopRepositoryMockForAuth) FindShopIDByAdminID(ctx context.Context, dbtx repositories.DBTX, adminID int) (int, error) {
	if m.FindShopIDByAdminIDFunc != nil {
		return m.FindShopIDByAdminIDFunc(ctx, dbtx, adminID)
	}
	panic("not implemented")
}

func (m *ShopRepositoryMockForAuth) FindAllShops(ctx context.Context, dbtx repositories.DBTX) ([]models.Shop, error) {
	panic("not implemented")
}

func (m *ShopRepositoryMockForAuth) UpdateShop(ctx context.Context, dbtx repositories.DBTX, shop *models.Shop) error {
	panic("not implemented")
}

func (m *ShopRepositoryMockForAuth) DeleteShop(ctx context.Context, dbtx repositories.DBTX, shopID int) error {
	panic("not implemented")
}

func (m *ShopRepositoryMockForAuth) FindShopByID(ctx context.Context, dbtx repositories.DBTX, shopID int) (*models.Shop, error) {
	panic("not implemented")
}

// OrderRepositoryMockForAuth - OrderRepositoryのモック実装（Auth用、DBTX対応）
type OrderRepositoryMockForAuth struct {
	UpdateUserIDByGuestTokenFunc func(ctx context.Context, dbtx repositories.DBTX, guestToken string, userID int) error
}

// NewOrderRepositoryMockForAuth モック実装を返す
func NewOrderRepositoryMockForAuth() *OrderRepositoryMockForAuth {
	return &OrderRepositoryMockForAuth{}
}

// インターフェース実装
func (m *OrderRepositoryMockForAuth) CreateOrder(ctx context.Context, dbtx repositories.DBTX, order *models.Order, items []models.OrderItem) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) UpdateUserIDByGuestToken(ctx context.Context, dbtx repositories.DBTX, guestToken string, userID int) error {
	if m.UpdateUserIDByGuestTokenFunc != nil {
		return m.UpdateUserIDByGuestTokenFunc(ctx, dbtx, guestToken, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindActiveUserOrders(ctx context.Context, dbtx repositories.DBTX, userID int) ([]repositories.OrderWithDetailsDB, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindItemsByOrderIDs(ctx context.Context, dbtx repositories.DBTX, orderIDs []int) (map[int][]models.ItemDetail, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindOrderByIDAndUser(ctx context.Context, dbtx repositories.DBTX, orderID int, userID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) CountWaitingOrders(ctx context.Context, dbtx repositories.DBTX, shopID int, orderDate time.Time) (int, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindShopOrdersByStatuses(ctx context.Context, dbtx repositories.DBTX, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindOrderByIDAndShopID(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) UpdateOrderStatus(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int, newStatus models.OrderStatus) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) DeleteOrderByIDAndShopID(ctx context.Context, dbtx repositories.DBTX, orderID int, shopID int) error {
	panic("not implemented")
}

// テスト定数
const (
	testEmail           = "test@example.com"
	testUserID          = 1
	testShopID          = 1
	testGuestOrderToken = "guest-token-123"
)

// setupTestEnv は.envファイルから環境変数を読み込み、SECRET_KEYを取得する
func setupTestEnv(t *testing.T) string {
	t.Helper()

	// .envファイルを読み込み（エラーは無視）
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		t.Skip("SECRET_KEY not set, skipping auth service test")
	}

	return secretKey
}

// TestAuthService_SignUp - ユーザー登録のテスト（将来のDBTX対応版）
func TestAuthService_SignUp(t *testing.T) {
	// .envファイルからSECRET_KEYを読み込み
	testSecret := setupTestEnv(t)

	tests := []struct {
		name             string
		req              models.AuthenticateRequest
		setupUserRepo    func(*UserRepositoryMockForAuth)
		setupOrderRepo   func(*OrderRepositoryMockForAuth)
		setupShopRepo    func(*ShopRepositoryMockForAuth)
		wantUserResponse models.UserResponse
		wantTokenEmpty   bool
		expectedErrCode  apperrors.ErrCode
	}{
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンなし）",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMockForAuth) {
				m.CreateUserFunc = func(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.CustomerRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:  func(m *ShopRepositoryMockForAuth) {},
			wantUserResponse: models.UserResponse{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.CustomerRole.String(),
			},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "異常系: ユーザー作成エラー",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMockForAuth) {
				m.CreateUserFunc = func(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
					return apperrors.Unknown.Wrap(nil, "ユーザー作成エラー")
				}
			},
			setupOrderRepo:   func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:    func(m *ShopRepositoryMockForAuth) {},
			wantUserResponse: models.UserResponse{},
			wantTokenEmpty:   true,
			expectedErrCode:  apperrors.Unknown,
		},
		{
			name: "正常系: 管理者ユーザー登録",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMockForAuth) {
				m.CreateUserFunc = func(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.AdminRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMockForAuth) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, dbtx repositories.DBTX, adminID int) (int, error) {
					return testShopID, nil
				}
			},
			wantUserResponse: models.UserResponse{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.AdminRole.String(),
			},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "異常系: 管理者ユーザーだがショップ情報取得エラー",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMockForAuth) {
				m.CreateUserFunc = func(ctx context.Context, dbtx repositories.DBTX, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.AdminRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMockForAuth) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, dbtx repositories.DBTX, adminID int) (int, error) {
					return 0, apperrors.NoData.Wrap(nil, "ショップが見つかりません")
				}
			},
			wantUserResponse: models.UserResponse{},
			wantTokenEmpty:   true,
			expectedErrCode:  apperrors.Unknown,
		},
		// NOTE: ゲストトークンありのテストケースは結合テスト（auth_service_integration_test.go）で実施
		// 理由: トランザクション処理が必要で、実際のDB接続が必要なため
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			userRepo := NewUserRepositoryMockForAuth()
			tt.setupUserRepo(userRepo)

			shopRepo := NewShopRepositoryMockForAuth()
			tt.setupShopRepo(shopRepo)

			orderRepo := NewOrderRepositoryMockForAuth()
			tt.setupOrderRepo(orderRepo)

			// サービス作成（単体テスト用 - トランザクションが必要な場合は結合テストで実施）
			mockDB := &sqlx.DB{} // 注意: これはトランザクションを使わないケースのみでテスト
			authService := services.NewAuthService(userRepo, shopRepo, orderRepo, mockDB)

			// テスト実行
			gotUser, gotToken, err := authService.SignUp(context.Background(), tt.req)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				testhelpers.AssertNoError(t, err)
			} else {
				testhelpers.AssertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if diff := cmp.Diff(tt.wantUserResponse, gotUser); diff != "" {
					t.Errorf("%s: user response mismatch (-want +got):\n%s", tt.name, diff)
				}

				if tt.wantTokenEmpty && gotToken != "" {
					t.Errorf("%s: expected empty token, got: %s", tt.name, gotToken)
				}

				if !tt.wantTokenEmpty && gotToken == "" {
					t.Errorf("%s: expected non-empty token, got empty", tt.name)
				}

				// トークンの検証（空でない場合）
				if !tt.wantTokenEmpty && gotToken != "" {
					token, parseErr := jwt.ParseWithClaims(gotToken, &models.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(testSecret), nil
					})
					if parseErr != nil {
						t.Errorf("%s: failed to parse token: %v", tt.name, parseErr)
					}

					if claims, ok := token.Claims.(*models.JwtCustomClaims); ok && token.Valid {
						if claims.UserID != tt.wantUserResponse.UserID {
							t.Errorf("%s: token UserID mismatch: want %d, got %d", tt.name, tt.wantUserResponse.UserID, claims.UserID)
						}
					} else {
						t.Errorf("%s: invalid token claims", tt.name)
					}
				}
			}
		})
	}
}
