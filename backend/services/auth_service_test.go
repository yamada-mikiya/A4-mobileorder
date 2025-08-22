package services_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/joho/godotenv"
)

// UserRepositoryMock - UserRepositoryのモック実装
type UserRepositoryMock struct {
	CreateUserFunc     func(ctx context.Context, user *models.User) error
	GetUserByEmailFunc func(ctx context.Context, email string) (models.User, error)
}

// NewUserRepositoryMock モック実装を返す
// NOTE: テストでモック関数を設定するため、具体的な型を返す
func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{}
}

// インターフェース実装
func (m *UserRepositoryMock) CreateUser(ctx context.Context, user *models.User) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	panic("not implemented")
}

func (m *UserRepositoryMock) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	panic("not implemented")
}

// ShopRepositoryMock - ShopRepositoryのモック実装
type ShopRepositoryMock struct {
	FindShopIDByAdminIDFunc func(ctx context.Context, adminID int) (int, error)
}

// NewShopRepositoryMock モック実装を返す
// NOTE: テストでモック関数を設定するため、具体的な型を返す
func NewShopRepositoryMock() *ShopRepositoryMock {
	return &ShopRepositoryMock{}
}

// インターフェース実装
func (m *ShopRepositoryMock) CreateShop(ctx context.Context, shop *models.Shop) error {
	panic("not implemented")
}

func (m *ShopRepositoryMock) FindShopIDByAdminID(ctx context.Context, adminID int) (int, error) {
	if m.FindShopIDByAdminIDFunc != nil {
		return m.FindShopIDByAdminIDFunc(ctx, adminID)
	}
	panic("not implemented")
}

func (m *ShopRepositoryMock) FindAllShops(ctx context.Context) ([]models.Shop, error) {
	panic("not implemented")
}

func (m *ShopRepositoryMock) UpdateShop(ctx context.Context, shop *models.Shop) error {
	panic("not implemented")
}

func (m *ShopRepositoryMock) DeleteShop(ctx context.Context, shopID int) error {
	panic("not implemented")
}

// OrderRepositoryMockForAuth - AuthService用のOrderRepositoryモック
type OrderRepositoryMockForAuth struct {
	UpdateUserIDByGuestTokenFunc func(ctx context.Context, guestToken string, userID int) error
}

// NewOrderRepositoryMockForAuth モック実装を返す
// NOTE: テストでモック関数を設定するため、具体的な型を返す
func NewOrderRepositoryMockForAuth() *OrderRepositoryMockForAuth {
	return &OrderRepositoryMockForAuth{}
}

// インターフェース実装
func (m *OrderRepositoryMockForAuth) CreateOrder(ctx context.Context, order *models.Order, items []models.OrderItem) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) UpdateUserIDByGuestToken(ctx context.Context, guestToken string, userID int) error {
	if m.UpdateUserIDByGuestTokenFunc != nil {
		return m.UpdateUserIDByGuestTokenFunc(ctx, guestToken, userID)
	}
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindActiveUserOrders(ctx context.Context, userID int) ([]repositories.OrderWithDetailsDB, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindItemsByOrderIDs(ctx context.Context, orderIDs []int) (map[int][]models.ItemDetail, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindOrderByIDAndUser(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) CountWaitingOrders(ctx context.Context, shopID int, orderDate time.Time) (int, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindShopOrdersByStatuses(ctx context.Context, shopID int, statuses []models.OrderStatus) ([]repositories.AdminOrderDBResult, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) FindOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) (*models.Order, error) {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) UpdateOrderStatus(ctx context.Context, orderID int, shopID int, newStatus models.OrderStatus) error {
	panic("not implemented")
}

func (m *OrderRepositoryMockForAuth) DeleteOrderByIDAndShopID(ctx context.Context, orderID int, shopID int) error {
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

func TestAuthService_SignUp(t *testing.T) {
	// .envファイルからSECRET_KEYを読み込み
	testSecret := setupTestEnv(t)

	tests := []struct {
		name             string
		req              models.AuthenticateRequest
		setupUserRepo    func(*UserRepositoryMock)
		setupOrderRepo   func(*OrderRepositoryMockForAuth)
		setupShopRepo    func(*ShopRepositoryMock)
		wantUserResponse models.UserResponse
		wantTokenEmpty   bool
		expectedErrCode  apperrors.ErrCode
	}{
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンなし）",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.CreateUserFunc = func(ctx context.Context, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.CustomerRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:  func(m *ShopRepositoryMock) {},
			wantUserResponse: models.UserResponse{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.CustomerRole.String(),
			},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "正常系: 新規ユーザー登録（ゲストトークンあり）",
			req: models.AuthenticateRequest{
				Email:           testEmail,
				GuestOrderToken: testGuestOrderToken,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.CreateUserFunc = func(ctx context.Context, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.CustomerRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {
				m.UpdateUserIDByGuestTokenFunc = func(ctx context.Context, guestToken string, userID int) error {
					return nil
				}
			},
			setupShopRepo: func(m *ShopRepositoryMock) {},
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
			setupUserRepo: func(m *UserRepositoryMock) {
				m.CreateUserFunc = func(ctx context.Context, user *models.User) error {
					return apperrors.Unknown.Wrap(nil, "ユーザー作成エラー")
				}
			},
			setupOrderRepo:   func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:    func(m *ShopRepositoryMock) {},
			wantUserResponse: models.UserResponse{},
			wantTokenEmpty:   true,
			expectedErrCode:  apperrors.Unknown,
		},
		{
			name: "正常系: 管理者ユーザー登録",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.CreateUserFunc = func(ctx context.Context, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.AdminRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
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
			setupUserRepo: func(m *UserRepositoryMock) {
				m.CreateUserFunc = func(ctx context.Context, user *models.User) error {
					user.UserID = testUserID
					user.Role = models.AdminRole
					return nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
					return 0, apperrors.NoData.Wrap(nil, "ショップが見つかりません")
				}
			},
			wantUserResponse: models.UserResponse{},
			wantTokenEmpty:   true,
			expectedErrCode:  apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			userRepo := NewUserRepositoryMock()
			tt.setupUserRepo(userRepo)

			shopRepo := NewShopRepositoryMock()
			tt.setupShopRepo(shopRepo)

			orderRepo := NewOrderRepositoryMockForAuth()
			tt.setupOrderRepo(orderRepo)

			// サービス作成 - ゲストトークンがある場合はFullTransactionが必要
			var mockTxm services.TransactionManager
			if tt.req.GuestOrderToken != "" {
				mockTxm = services.NewMockTransactionManagerFull(userRepo, orderRepo)
			} else {
				mockTxm = services.NewMockTransactionManager(orderRepo)
			}
			authService := services.NewAuthServiceForTest(userRepo, shopRepo, orderRepo, mockTxm)

			// テスト実行
			gotUser, gotToken, err := authService.SignUp(context.Background(), tt.req) // エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
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

func TestAuthService_LogIn(t *testing.T) {
	// .envファイルからSECRET_KEYを読み込み
	testSecret := setupTestEnv(t)

	tests := []struct {
		name            string
		req             models.AuthenticateRequest
		setupUserRepo   func(*UserRepositoryMock)
		setupOrderRepo  func(*OrderRepositoryMockForAuth)
		setupShopRepo   func(*ShopRepositoryMock)
		wantTokenEmpty  bool
		expectedErrCode apperrors.ErrCode
	}{
		{
			name: "正常系: ログイン成功（ゲストトークンなし）",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						UserID: testUserID,
						Email:  testEmail,
						Role:   models.CustomerRole,
					}, nil
				}
			},
			setupOrderRepo:  func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:   func(m *ShopRepositoryMock) {},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "正常系: ログイン成功（ゲストトークンあり）",
			req: models.AuthenticateRequest{
				Email:           testEmail,
				GuestOrderToken: testGuestOrderToken,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						UserID: testUserID,
						Email:  testEmail,
						Role:   models.CustomerRole,
					}, nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {
				m.UpdateUserIDByGuestTokenFunc = func(ctx context.Context, guestToken string, userID int) error {
					return nil
				}
			},
			setupShopRepo:   func(m *ShopRepositoryMock) {},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "異常系: ユーザーが見つからない（NoDataエラー）",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{}, apperrors.NoData.Wrap(nil, "ユーザーが見つかりません")
				}
			},
			setupOrderRepo:  func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:   func(m *ShopRepositoryMock) {},
			wantTokenEmpty:  true,
			expectedErrCode: apperrors.Unauthorized,
		},
		{
			name: "異常系: その他のデータベースエラー",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{}, apperrors.Unknown.Wrap(nil, "データベースエラー")
				}
			},
			setupOrderRepo:  func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo:   func(m *ShopRepositoryMock) {},
			wantTokenEmpty:  true,
			expectedErrCode: apperrors.Unknown,
		},
		{
			name: "正常系: 管理者ユーザーのログイン",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						UserID: testUserID,
						Email:  testEmail,
						Role:   models.AdminRole,
					}, nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
					return testShopID, nil
				}
			},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "異常系: 管理者ユーザーだがショップ情報取得エラー",
			req: models.AuthenticateRequest{
				Email: testEmail,
			},
			setupUserRepo: func(m *UserRepositoryMock) {
				m.GetUserByEmailFunc = func(ctx context.Context, email string) (models.User, error) {
					return models.User{
						UserID: testUserID,
						Email:  testEmail,
						Role:   models.AdminRole,
					}, nil
				}
			},
			setupOrderRepo: func(m *OrderRepositoryMockForAuth) {},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
					return 0, apperrors.NoData.Wrap(nil, "ショップが見つかりません")
				}
			},
			wantTokenEmpty:  true,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			userRepo := NewUserRepositoryMock()
			tt.setupUserRepo(userRepo)

			shopRepo := NewShopRepositoryMock()
			tt.setupShopRepo(shopRepo)

			orderRepo := NewOrderRepositoryMockForAuth()
			tt.setupOrderRepo(orderRepo)

			// サービス作成 - ゲストトークンがある場合はFullTransactionが必要
			var mockTxm services.TransactionManager
			if tt.req.GuestOrderToken != "" {
				mockTxm = services.NewMockTransactionManagerFull(userRepo, orderRepo)
			} else {
				mockTxm = services.NewMockTransactionManager(orderRepo)
			}
			authService := services.NewAuthServiceForTest(userRepo, shopRepo, orderRepo, mockTxm)

			// テスト実行
			gotUser, gotToken, err := authService.LogIn(context.Background(), tt.req) // エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				// UserResponseのアサーション - テストケースに応じてRoleを設定
				expectedRole := models.CustomerRole.String()
				if tt.name == "正常系: 管理者ユーザーのログイン" {
					expectedRole = models.AdminRole.String()
				}
				expectedUser := models.UserResponse{
					UserID: testUserID,
					Email:  testEmail,
					Role:   expectedRole,
				}
				if diff := cmp.Diff(expectedUser, gotUser); diff != "" {
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
						if claims.UserID != testUserID {
							t.Errorf("%s: token UserID mismatch: want %d, got %d", tt.name, testUserID, claims.UserID)
						}
					} else {
						t.Errorf("%s: invalid token claims", tt.name)
					}
				}
			}
		})
	}
}

func TestAuthService_createToken(t *testing.T) {
	// .envファイルからSECRET_KEYを読み込み
	testSecret := setupTestEnv(t)

	tests := []struct {
		name            string
		user            models.User
		setupShopRepo   func(*ShopRepositoryMock)
		wantTokenEmpty  bool
		expectedErrCode apperrors.ErrCode
	}{
		{
			name: "正常系: 一般ユーザーのトークン生成",
			user: models.User{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.CustomerRole,
			},
			setupShopRepo:   func(m *ShopRepositoryMock) {},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "正常系: 管理者ユーザーのトークン生成",
			user: models.User{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.AdminRole,
			},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
					return testShopID, nil
				}
			},
			wantTokenEmpty:  false,
			expectedErrCode: "",
		},
		{
			name: "異常系: 管理者ユーザーだがショップ情報取得エラー",
			user: models.User{
				UserID: testUserID,
				Email:  testEmail,
				Role:   models.AdminRole,
			},
			setupShopRepo: func(m *ShopRepositoryMock) {
				m.FindShopIDByAdminIDFunc = func(ctx context.Context, adminID int) (int, error) {
					return 0, apperrors.NoData.Wrap(nil, "ショップが見つかりません")
				}
			},
			wantTokenEmpty:  true,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックセットアップ
			userRepo := NewUserRepositoryMock()
			shopRepo := NewShopRepositoryMock()
			tt.setupShopRepo(shopRepo)
			orderRepo := NewOrderRepositoryMockForAuth()

			// サービス作成（createTokenはprivateメソッドなのでリフレクションまたは他の手段が必要）
			// 今回はSignUpまたはLogIn経由でテストするか、createTokenを公開メソッドにする必要がある
			// ここでは直接的にはテストできないため、SignUpやLogInのテスト内でカバーする

			// 代替案: createTokenメソッドのロジックを間接的にテスト
			mockTxm := services.NewMockTransactionManagerFull(userRepo, orderRepo)
			authService := services.NewAuthServiceForTest(userRepo, shopRepo, orderRepo, mockTxm)

			// SignUpメソッド経由でcreateTokenをテスト
			req := models.AuthenticateRequest{Email: tt.user.Email}

			userRepo.CreateUserFunc = func(ctx context.Context, user *models.User) error {
				user.UserID = tt.user.UserID
				user.Role = tt.user.Role
				return nil
			}

			_, gotToken, err := authService.SignUp(context.Background(), req)

			// エラーアサーション
			if tt.expectedErrCode == "" {
				assertNoError(t, err)
			} else {
				assertAppError(t, err, tt.expectedErrCode)
			}

			// 正常系の場合のアサーション
			if tt.expectedErrCode == "" {
				if tt.wantTokenEmpty && gotToken != "" {
					t.Errorf("%s: expected empty token, got: %s", tt.name, gotToken)
				}

				if !tt.wantTokenEmpty && gotToken == "" {
					t.Errorf("%s: expected non-empty token, got empty", tt.name)
				}

				// トークンの内容検証
				if !tt.wantTokenEmpty && gotToken != "" {
					token, parseErr := jwt.ParseWithClaims(gotToken, &models.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(testSecret), nil
					})
					if parseErr != nil {
						t.Errorf("%s: failed to parse token: %v", tt.name, parseErr)
						return
					}

					if claims, ok := token.Claims.(*models.JwtCustomClaims); ok && token.Valid {
						if claims.UserID != tt.user.UserID {
							t.Errorf("%s: token UserID mismatch: want %d, got %d", tt.name, tt.user.UserID, claims.UserID)
						}
						if claims.Role != tt.user.Role {
							t.Errorf("%s: token Role mismatch: want %v, got %v", tt.name, tt.user.Role, claims.Role)
						}

						// 管理者の場合、ShopIDが設定されているかチェック
						if tt.user.Role == models.AdminRole {
							if claims.ShopID == nil {
								t.Errorf("%s: expected ShopID to be set for admin user", tt.name)
							} else if *claims.ShopID != testShopID {
								t.Errorf("%s: token ShopID mismatch: want %d, got %d", tt.name, testShopID, *claims.ShopID)
							}
						} else {
							if claims.ShopID != nil {
								t.Errorf("%s: expected ShopID to be nil for non-admin user", tt.name)
							}
						}

						// 有効期限のチェック
						if claims.ExpiresAt == nil {
							t.Errorf("%s: expected ExpiresAt to be set", tt.name)
						} else {
							expectedExpiry := time.Now().Add(time.Hour * 72)
							actualExpiry := claims.ExpiresAt.Time
							// 許容誤差: 1分
							if actualExpiry.Before(expectedExpiry.Add(-time.Minute)) || actualExpiry.After(expectedExpiry.Add(time.Minute)) {
								t.Errorf("%s: token expiry time out of expected range", tt.name)
							}
						}
					} else {
						t.Errorf("%s: invalid token claims", tt.name)
					}
				}
			}
		})
	}
}
