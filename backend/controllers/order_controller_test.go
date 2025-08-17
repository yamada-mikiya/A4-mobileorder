package controllers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// MockOrderService は OrderServicer インターフェースのモック実装です
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) CreateOrder(ctx context.Context, shopID int, reqItem []models.OrderItemRequest) (*models.Order, error) {
	args := m.Called(ctx, shopID, reqItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) CreateAuthenticatedOrder(ctx context.Context, userID int, shopID int, items []models.OrderItemRequest) (*models.Order, error) {
	args := m.Called(ctx, userID, shopID, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderService) GetUserOrders(ctx context.Context, userID int) ([]models.OrderListResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.OrderListResponse), args.Error(1)
}

func (m *MockOrderService) GetOrderStatus(ctx context.Context, userID int, orderID int) (*models.OrderStatusResponse, error) {
	args := m.Called(ctx, userID, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OrderStatusResponse), args.Error(1)
}

// createTestContextForOrder はOrder用のEchoコンテキストを作成します
func createTestContextForOrder(method, path string, body string, pathParams map[string]string, token *jwt.Token) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
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

// TestOrderController_CreateAuthenticatedOrderHandler のテストケース
func TestOrderController_CreateAuthenticatedOrderHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		pathParams       map[string]string
		setupMock        func() (*MockOrderService, *MockValidator)
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: 認証済みユーザーの注文作成成功",
			requestBody: `{"items":[{"item_id":1,"quantity":2},{"item_id":2,"quantity":1}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				order := &models.Order{
					OrderID:     123,
					UserID:      sql.NullInt64{Int64: 1, Valid: true},
					ShopID:      1,
					TotalAmount: 1500.0,
					Status:      models.Cooking,
					OrderDate:   time.Now(),
				}

				mockValidator.On("Validate", mock.Anything).Return(nil)
				mockService.On("CreateAuthenticatedOrder", mock.Anything, 1, 1, mock.MatchedBy(func(items []models.OrderItemRequest) bool {
					return len(items) == 2 && items[0].ItemID == 1 && items[0].Quantity == 2
				})).Return(order, nil)

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthenticatedOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, uint(123), response.OrderID)
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"items":}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー",
			requestBody: `{"items":[]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockValidator.On("Validate", mock.Anything).Return(apperrors.ValidationFailed.Wrap(nil, "注文には少なくとも1つの商品が必要です"))

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 店舗IDの形式が不正",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "invalid"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				// バリデーションは店舗IDチェック前に呼ばれるため、成功させる
				mockValidator.On("Validate", mock.Anything).Return(nil)
				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:        "異常系: 認証トークンなし",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				// バリデーションは認証チェック前に呼ばれるため、成功させる
				mockValidator.On("Validate", mock.Anything).Return(nil)
				return mockService, mockValidator
			},
			setupToken:     func() *jwt.Token { return nil },
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			expectedCode:   apperrors.Unauthorized,
		},
		{
			name:        "異常系: サービス層でエラー発生",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockValidator.On("Validate", mock.Anything).Return(nil)
				mockService.On("CreateAuthenticatedOrder", mock.Anything, 1, 1, mock.Anything).Return(
					nil, apperrors.Unknown.Wrap(nil, "注文作成に失敗しました"))

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService, mockValidator := tt.setupMock()
			defer mockService.AssertExpectations(t)
			defer mockValidator.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewOrderController(mockService, mockValidator)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodPost, "/shops/1/orders", tt.requestBody, tt.pathParams, token)

			// テスト実行
			err := controller.CreateAuthenticatedOrderHandler(c)

			// エラーアサーション
			if tt.expectError {
				assert.Error(t, err)
				var appErr *apperrors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedCode, appErr.ErrCode)
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

// TestOrderController_CreateGuestOrderHandler のテストケース
func TestOrderController_CreateGuestOrderHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		pathParams       map[string]string
		setupMock        func() (*MockOrderService, *MockValidator)
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: ゲストユーザーの注文作成成功",
			requestBody: `{"items":[{"item_id":1,"quantity":2},{"item_id":2,"quantity":1}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				order := &models.Order{
					OrderID:         456,
					ShopID:          1,
					TotalAmount:     1500.0,
					Status:          models.Cooking,
					GuestOrderToken: sql.NullString{String: "guest-token-123", Valid: true},
					OrderDate:       time.Now(),
				}

				mockValidator.On("Validate", mock.Anything).Return(nil)
				mockService.On("CreateOrder", mock.Anything, 1, mock.MatchedBy(func(items []models.OrderItemRequest) bool {
					return len(items) == 2 && items[0].ItemID == 1 && items[0].Quantity == 2
				})).Return(order, nil)

				return mockService, mockValidator
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.CreateOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 456, response.OrderID)
				assert.Equal(t, "guest-token-123", response.GuestOrderToken)
				assert.Contains(t, response.Message, "Order created successfully")
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"items":}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				return mockService, mockValidator
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー",
			requestBody: `{"items":[]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockValidator.On("Validate", mock.Anything).Return(apperrors.ValidationFailed.Wrap(nil, "注文には少なくとも1つの商品が必要です"))

				return mockService, mockValidator
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 店舗IDの形式が不正",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "invalid"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				// バリデーションは店舗IDチェック前に呼ばれるため、成功させる
				mockValidator.On("Validate", mock.Anything).Return(nil)
				return mockService, mockValidator
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:        "異常系: サービス層でエラー発生",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockValidator.On("Validate", mock.Anything).Return(nil)
				mockService.On("CreateOrder", mock.Anything, 1, mock.Anything).Return(
					nil, apperrors.Unknown.Wrap(nil, "注文作成に失敗しました"))

				return mockService, mockValidator
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService, mockValidator := tt.setupMock()
			defer mockService.AssertExpectations(t)
			defer mockValidator.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewOrderController(mockService, mockValidator)

			// テストコンテキストの作成
			c, rec := createTestContextForOrder(http.MethodPost, "/shops/1/guest-orders", tt.requestBody, tt.pathParams, nil)

			// テスト実行
			err := controller.CreateGuestOrderHandler(c)

			// エラーアサーション
			if tt.expectError {
				assert.Error(t, err)
				var appErr *apperrors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedCode, appErr.ErrCode)
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

// TestOrderController_GetOrderListHandler のテストケース
func TestOrderController_GetOrderListHandler(t *testing.T) {
	tests := []struct {
		name             string
		setupMock        func() (*MockOrderService, *MockValidator)
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 注文一覧取得成功",
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				orderList := []models.OrderListResponse{
					{
						OrderID:      1,
						ShopName:     "テストショップ",
						Location:     "東京都渋谷区",
						OrderDate:    time.Date(2025, 8, 17, 12, 0, 0, 0, time.UTC),
						TotalAmount:  1500.0,
						Status:       "cooking",
						WaitingCount: 3,
						Items: []models.ItemDetail{
							{ItemName: "コーヒー", Quantity: 2},
							{ItemName: "サンドイッチ", Quantity: 1},
						},
					},
				}

				mockService.On("GetUserOrders", mock.Anything, 1).Return(orderList, nil)

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response []models.OrderListResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 1)
				assert.Equal(t, 1, response[0].OrderID)
				assert.Equal(t, "テストショップ", response[0].ShopName)
				assert.Equal(t, "cooking", response[0].Status)
				assert.Len(t, response[0].Items, 2)
			},
		},
		{
			name: "正常系: 注文一覧が空の場合",
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				orderList := []models.OrderListResponse{}
				mockService.On("GetUserOrders", mock.Anything, 1).Return(orderList, nil)

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response []models.OrderListResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name: "異常系: 認証トークンなし",
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				return mockService, mockValidator
			},
			setupToken:     func() *jwt.Token { return nil },
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			expectedCode:   apperrors.Unauthorized,
		},
		{
			name: "異常系: サービス層でエラー発生",
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockService.On("GetUserOrders", mock.Anything, 1).Return(
					[]models.OrderListResponse{}, apperrors.GetDataFailed.Wrap(nil, "注文一覧の取得に失敗しました"))

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.GetDataFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService, mockValidator := tt.setupMock()
			defer mockService.AssertExpectations(t)
			defer mockValidator.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewOrderController(mockService, mockValidator)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodGet, "/orders", "", nil, token)

			// テスト実行
			err := controller.GetOrderListHandler(c)

			// エラーアサーション
			if tt.expectError {
				assert.Error(t, err)
				var appErr *apperrors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedCode, appErr.ErrCode)
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

// TestOrderController_GetOrderStatusHandler のテストケース
func TestOrderController_GetOrderStatusHandler(t *testing.T) {
	tests := []struct {
		name             string
		pathParams       map[string]string
		setupMock        func() (*MockOrderService, *MockValidator)
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "正常系: 注文ステータス取得成功（調理中）",
			pathParams: map[string]string{"order_id": "123"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				orderStatus := &models.OrderStatusResponse{
					OrderID:      123,
					Status:       "cooking",
					WaitingCount: 3,
				}

				mockService.On("GetOrderStatus", mock.Anything, 1, 123).Return(orderStatus, nil)

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.OrderStatusResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 123, response.OrderID)
				assert.Equal(t, "cooking", response.Status)
				assert.Equal(t, 3, response.WaitingCount)
			},
		},
		{
			name:       "正常系: 注文ステータス取得成功（完了）",
			pathParams: map[string]string{"order_id": "456"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				orderStatus := &models.OrderStatusResponse{
					OrderID:      456,
					Status:       "completed",
					WaitingCount: 0, // 完了時は待ち人数0
				}

				mockService.On("GetOrderStatus", mock.Anything, 1, 456).Return(orderStatus, nil)

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.OrderStatusResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 456, response.OrderID)
				assert.Equal(t, "completed", response.Status)
				assert.Equal(t, 0, response.WaitingCount)
			},
		},
		{
			name:       "異常系: 注文IDの形式が不正",
			pathParams: map[string]string{"order_id": "invalid"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:       "異常系: 認証トークンなし",
			pathParams: map[string]string{"order_id": "123"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)
				return mockService, mockValidator
			},
			setupToken:     func() *jwt.Token { return nil },
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			expectedCode:   apperrors.Unauthorized,
		},
		{
			name:       "異常系: 注文が見つからない",
			pathParams: map[string]string{"order_id": "999"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockService.On("GetOrderStatus", mock.Anything, 1, 999).Return(
					nil, apperrors.NoData.Wrap(nil, "注文が見つかりません"))

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			expectedCode:   apperrors.NoData,
		},
		{
			name:       "異常系: サービス層で内部エラー発生",
			pathParams: map[string]string{"order_id": "123"},
			setupMock: func() (*MockOrderService, *MockValidator) {
				mockService := new(MockOrderService)
				mockValidator := new(MockValidator)

				mockService.On("GetOrderStatus", mock.Anything, 1, 123).Return(
					nil, apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))

				return mockService, mockValidator
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService, mockValidator := tt.setupMock()
			defer mockService.AssertExpectations(t)
			defer mockValidator.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewOrderController(mockService, mockValidator)

			// テストコンテキストの作成
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodGet, "/orders/123/status", "", tt.pathParams, token)

			// テスト実行
			err := controller.GetOrderStatusHandler(c)

			// エラーアサーション
			if tt.expectError {
				assert.Error(t, err)
				var appErr *apperrors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedCode, appErr.ErrCode)
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
