package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAdminService は AdminServicer インターフェースのモック実装です
type MockAdminService struct {
	mock.Mock
}

func (m *MockAdminService) GetCookingOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {
	args := m.Called(ctx, shopID)
	return args.Get(0).([]models.AdminOrderResponse), args.Error(1)
}

func (m *MockAdminService) GetCompletedOrders(ctx context.Context, shopID int) ([]models.AdminOrderResponse, error) {
	args := m.Called(ctx, shopID)
	return args.Get(0).([]models.AdminOrderResponse), args.Error(1)
}

func (m *MockAdminService) UpdateOrderStatus(ctx context.Context, adminShopID int, targetOrderID int) error {
	args := m.Called(ctx, adminShopID, targetOrderID)
	return args.Error(0)
}

func (m *MockAdminService) DeleteOrder(ctx context.Context, adminShopID int, targetOrderID int) error {
	args := m.Called(ctx, adminShopID, targetOrderID)
	return args.Error(0)
}

// createTestToken はテスト用のJWTトークンを作成します
func createTestToken(userID int, role models.UserRole, shopID *int) *jwt.Token {
	claims := &models.JwtCustomClaims{
		UserID: userID,
		Role:   role,
		ShopID: shopID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": jwt.SigningMethodHS256.Alg(),
		},
		Claims: claims,
		Method: jwt.SigningMethodHS256,
		Valid:  true,
	}
	return token
}

// createTestContext はテスト用のEchoコンテキストを作成します
func createTestContext(method, path string, pathParams map[string]string, token *jwt.Token) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// パスパラメータを設定
	if pathParams != nil {
		var names []string
		var values []string
		for name, value := range pathParams {
			names = append(names, name)
			values = append(values, value)
		}
		c.SetParamNames(names...)
		c.SetParamValues(values...)
	}

	// JWTトークンを設定
	if token != nil {
		c.Set("user", token)
	}

	return c, rec
}

// テストデータ作成ヘルパー
func createSampleAdminOrderResponse() []models.AdminOrderResponse {
	return []models.AdminOrderResponse{
		{
			OrderID:       1,
			CustomerEmail: stringPtr("test@example.com"),
			OrderDate:     time.Date(2025, 8, 16, 12, 0, 0, 0, time.UTC),
			TotalAmount:   1500.0,
			Status:        "cooking",
			Items: []models.ItemDetail{
				{
					ItemName: "コーヒー",
					Quantity: 2,
				},
			},
		},
	}
}

func stringPtr(s string) *string {
	return &s
}

// TestAdminController_GetCookingOrdersHandler のテストケース
func TestAdminController_GetCookingOrdersHandler(t *testing.T) {
	tests := []struct {
		name             string
		shopID           string
		setupMock        func() *MockAdminService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:   "正常系: 調理中の注文一覧取得成功",
			shopID: "1",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockOrders := createSampleAdminOrderResponse()
				mockService.On("GetCookingOrders", mock.Anything, 1).Return(mockOrders, nil)
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response []models.AdminOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, 1, response[0].OrderID)
			},
		},
		{
			name:   "異常系: 店舗IDの形式が不正",
			shopID: "invalid",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.BadParam,
		},
		{
			name:   "異常系: トークンなし",
			shopID: "1",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				return nil
			},
			expectError:  true,
			expectedCode: apperrors.Unauthorized,
		},
		{
			name:   "異常系: 店舗IDが異なる管理者",
			shopID: "1",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 2 // 異なる店舗ID
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.Forbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAdminController(mockService)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContext(
				http.MethodGet,
				"/admin/shops/"+tt.shopID+"/orders/cooking",
				map[string]string{"shop_id": tt.shopID},
				token,
			)

			// ハンドラーの実行
			err := controller.GetCookingOrdersHandler(c)

			// 結果の検証
			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.validateResponse != nil {
					tt.validateResponse(t, rec)
				}
			}
		})
	}
}

// TestAdminController_GetCompletedOrdersHandler のテストケース
func TestAdminController_GetCompletedOrdersHandler(t *testing.T) {
	tests := []struct {
		name             string
		shopID           string
		setupMock        func() *MockAdminService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:   "正常系: 完了済み注文一覧取得成功",
			shopID: "1",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockOrders := []models.AdminOrderResponse{
					{
						OrderID:       2,
						CustomerEmail: stringPtr("completed@example.com"),
						OrderDate:     time.Date(2025, 8, 16, 10, 0, 0, 0, time.UTC),
						TotalAmount:   2000.0,
						Status:        "completed",
						Items: []models.ItemDetail{
							{
								ItemName: "ラテ",
								Quantity: 1,
							},
						},
					},
				}
				mockService.On("GetCompletedOrders", mock.Anything, 1).Return(mockOrders, nil)
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response []models.AdminOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, 2, response[0].OrderID)
				assert.Equal(t, "completed", response[0].Status)
			},
		},
		{
			name:   "正常系: 完了済み注文が空の場合",
			shopID: "1",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockOrders := []models.AdminOrderResponse{}
				mockService.On("GetCompletedOrders", mock.Anything, 1).Return(mockOrders, nil)
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response []models.AdminOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name:   "異常系: 店舗IDの形式が不正",
			shopID: "invalid",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.BadParam,
		},
		{
			name:   "異常系: トークンなし",
			shopID: "1",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				return nil
			},
			expectError:  true,
			expectedCode: apperrors.Unauthorized,
		},
		{
			name:   "異常系: 店舗IDが異なる管理者",
			shopID: "1",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 2 // 異なる店舗ID
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.Forbidden,
		},
		{
			name:   "異常系: サービス層でエラー発生",
			shopID: "1",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockService.On("GetCompletedOrders", mock.Anything, 1).Return([]models.AdminOrderResponse{}, apperrors.GetDataFailed.Wrap(nil, "データベースエラー"))
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.GetDataFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAdminController(mockService)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContext(
				http.MethodGet,
				"/admin/shops/"+tt.shopID+"/orders/completed",
				map[string]string{"shop_id": tt.shopID},
				token,
			)

			// ハンドラーの実行
			err := controller.GetCompletedOrdersHandler(c)

			// 結果の検証
			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.validateResponse != nil {
					tt.validateResponse(t, rec)
				}
			}
		})
	}
}

// TestAdminController_UpdateOrderStatusHandler のテストケース
func TestAdminController_UpdateOrderStatusHandler(t *testing.T) {
	tests := []struct {
		name             string
		orderID          string
		setupMock        func() *MockAdminService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "正常系: 注文ステータス更新成功",
			orderID: "123",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockService.On("UpdateOrderStatus", mock.Anything, 1, 123).Return(nil)
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "注文ステータスを更新しました。", response["message"])
			},
		},
		{
			name:    "異常系: 店舗IDがnilの管理者",
			orderID: "123",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.AdminRole, nil)
			},
			expectError:  true,
			expectedCode: apperrors.Forbidden,
		},
		{
			name:    "異常系: 注文IDの形式が不正",
			orderID: "invalid",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.BadParam,
		},
		{
			name:    "異常系: サービス層でエラー発生",
			orderID: "123",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockService.On("UpdateOrderStatus", mock.Anything, 1, 123).Return(apperrors.NoData.Wrap(nil, "注文が見つかりません"))
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAdminController(mockService)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContext(
				http.MethodPatch,
				"/admin/orders/"+tt.orderID+"/status",
				map[string]string{"order_id": tt.orderID},
				token,
			)

			// ハンドラーの実行
			err := controller.UpdateOrderStatusHandler(c)

			// 結果の検証
			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.validateResponse != nil {
					tt.validateResponse(t, rec)
				}
			}
		})
	}
}

// TestAdminController_DeleteOrderHandler のテストケース
func TestAdminController_DeleteOrderHandler(t *testing.T) {
	tests := []struct {
		name             string
		orderID          string
		setupMock        func() *MockAdminService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "正常系: 注文削除成功",
			orderID: "123",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockService.On("DeleteOrder", mock.Anything, 1, 123).Return(nil)
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "注文を削除しました。", response["message"])
			},
		},
		{
			name:    "異常系: サービス層で注文が見つからない",
			orderID: "123",
			setupMock: func() *MockAdminService {
				mockService := new(MockAdminService)
				mockService.On("DeleteOrder", mock.Anything, 1, 123).Return(apperrors.NoData.Wrap(nil, "注文が見つかりません"))
				return mockService
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.NoData,
		},
		{
			name:    "異常系: 注文IDの形式が不正",
			orderID: "invalid",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				adminShopID := 1
				return createTestToken(1, models.AdminRole, &adminShopID)
			},
			expectError:  true,
			expectedCode: apperrors.BadParam,
		},
		{
			name:    "異常系: 店舗IDがnilの管理者",
			orderID: "123",
			setupMock: func() *MockAdminService {
				return new(MockAdminService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.AdminRole, nil)
			},
			expectError:  true,
			expectedCode: apperrors.Forbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAdminController(mockService)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContext(
				http.MethodDelete,
				"/admin/orders/"+tt.orderID+"/delete",
				map[string]string{"order_id": tt.orderID},
				token,
			)

			// ハンドラーの実行
			err := controller.DeleteOrderHandler(c)

			// 結果の検証
			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
				if tt.validateResponse != nil {
					tt.validateResponse(t, rec)
				}
			}
		})
	}
}
