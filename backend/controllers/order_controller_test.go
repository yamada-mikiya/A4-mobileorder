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
		setupMock        func() *MockOrderService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: 認証済みユーザーの注文作成成功",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)

				order := &models.Order{
					OrderID:     1,
					UserID:      sql.NullInt64{Int64: 1, Valid: true},
					ShopID:      1,
					Status:      models.Cooking,
					TotalAmount: 2000,
					OrderDate:   time.Now(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				mockService.On("CreateAuthenticatedOrder", mock.Anything, 1, 1, mock.MatchedBy(func(items []models.OrderItemRequest) bool {
					return len(items) == 1 && items[0].ItemID == 1 && items[0].Quantity == 2
				})).Return(order, nil)

				return mockService
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
				assert.Equal(t, uint(1), response.OrderID)
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"items":}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー（空の商品リスト）",
			requestBody: `{"items":[]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 不正な店舗ID",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "invalid"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:        "異常系: サービス層でエラー",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				mockService.On("CreateAuthenticatedOrder", mock.Anything, 1, 1, mock.Anything).Return(
					nil, apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))
				return mockService
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
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			controller := controllers.NewOrderController(mockService)
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodPost, "/shops/1/orders", tt.requestBody, tt.pathParams, token)

			err := controller.CreateAuthenticatedOrderHandler(c)

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
		setupMock        func() *MockOrderService
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: ゲストユーザーの注文作成成功",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)

				order := &models.Order{
					OrderID:         1,
					UserID:          sql.NullInt64{Valid: false},
					ShopID:          1,
					Status:          models.Cooking,
					TotalAmount:     2000,
					GuestOrderToken: sql.NullString{String: "15ff4999-2cfd-41f3-b744-926e7c5c7a0e", Valid: true},
					OrderDate:       time.Now(),
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}

				mockService.On("CreateOrder", mock.Anything, 1, mock.MatchedBy(func(items []models.OrderItemRequest) bool {
					return len(items) == 1 && items[0].ItemID == 1 && items[0].Quantity == 2
				})).Return(order, nil)

				return mockService
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.CreateOrderResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 1, response.OrderID)
				assert.Equal(t, "15ff4999-2cfd-41f3-b744-926e7c5c7a0e", response.GuestOrderToken)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"items":}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー（空の商品リスト）",
			requestBody: `{"items":[]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 不正な店舗ID",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "invalid"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:        "異常系: サービス層でエラー",
			requestBody: `{"items":[{"item_id":1,"quantity":2}]}`,
			pathParams:  map[string]string{"shop_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				mockService.On("CreateOrder", mock.Anything, 1, mock.Anything).Return(
					nil, apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))
				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			controller := controllers.NewOrderController(mockService)
			c, rec := createTestContextForOrder(http.MethodPost, "/shops/1/guest/orders", tt.requestBody, tt.pathParams, nil)

			err := controller.CreateGuestOrderHandler(c)

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
		setupMock        func() *MockOrderService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 注文一覧取得成功",
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)

				orders := []models.OrderListResponse{
					{
						OrderID:      1,
						ShopName:     "テスト店舗",
						Location:     "東京都渋谷区",
						OrderDate:    time.Now(),
						TotalAmount:  2000,
						Status:       "cooking",
						WaitingCount: 3,
						Items: []models.ItemDetail{
							{ItemName: "テスト商品", Quantity: 2},
						},
					},
					{
						OrderID:      2,
						ShopName:     "テスト店舗2",
						Location:     "東京都新宿区",
						OrderDate:    time.Now().Add(-24 * time.Hour),
						TotalAmount:  1500,
						Status:       "completed",
						WaitingCount: 0,
						Items: []models.ItemDetail{
							{ItemName: "テスト商品2", Quantity: 1},
						},
					},
				}

				mockService.On("GetUserOrders", mock.Anything, 1).Return(orders, nil)
				return mockService
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
				assert.Len(t, response, 2)
				assert.Equal(t, 1, response[0].OrderID)
				assert.Equal(t, "テスト店舗", response[0].ShopName)
			},
		},
		{
			name: "正常系: 空の注文一覧",
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				orders := []models.OrderListResponse{}
				mockService.On("GetUserOrders", mock.Anything, 1).Return(orders, nil)
				return mockService
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
			name: "異常系: サービス層でエラー",
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				mockService.On("GetUserOrders", mock.Anything, 1).Return(
					[]models.OrderListResponse{}, apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))
				return mockService
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
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			controller := controllers.NewOrderController(mockService)
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodGet, "/orders", "", nil, token)

			err := controller.GetOrderListHandler(c)

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
		setupMock        func() *MockOrderService
		setupToken       func() *jwt.Token
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "正常系: 注文ステータス取得成功",
			pathParams: map[string]string{"order_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)

				orderStatus := &models.OrderStatusResponse{
					OrderID:      1,
					Status:       "cooking",
					WaitingCount: 2,
				}

				mockService.On("GetOrderStatus", mock.Anything, 1, 1).Return(orderStatus, nil)
				return mockService
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
				assert.Equal(t, 1, response.OrderID)
				assert.Equal(t, "cooking", response.Status)
				assert.Equal(t, 2, response.WaitingCount)
			},
		},
		{
			name:       "異常系: 不正な注文ID",
			pathParams: map[string]string{"order_id": "invalid"},
			setupMock: func() *MockOrderService {
				return new(MockOrderService)
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.BadParam,
		},
		{
			name:       "異常系: 注文が見つからない",
			pathParams: map[string]string{"order_id": "999"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				mockService.On("GetOrderStatus", mock.Anything, 1, 999).Return(
					nil, apperrors.NoData.Wrap(nil, "注文が見つかりません"))
				return mockService
			},
			setupToken: func() *jwt.Token {
				return createTestToken(1, models.CustomerRole, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			expectedCode:   apperrors.NoData,
		},
		{
			name:       "異常系: サービス層でエラー",
			pathParams: map[string]string{"order_id": "1"},
			setupMock: func() *MockOrderService {
				mockService := new(MockOrderService)
				mockService.On("GetOrderStatus", mock.Anything, 1, 1).Return(
					nil, apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))
				return mockService
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
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			controller := controllers.NewOrderController(mockService)
			token := tt.setupToken()
			c, rec := createTestContextForOrder(http.MethodGet, "/orders/1", "", tt.pathParams, token)

			err := controller.GetOrderStatusHandler(c)

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
