package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService は AuthServicer インターフェースのモック実装です
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.UserResponse), args.String(1), args.Error(2)
}

func (m *MockAuthService) LogIn(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.UserResponse), args.String(1), args.Error(2)
}

// createTestContextForAuth は認証関連のテスト用のEchoコンテキストを作成します
func createTestContextForAuth(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
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
	return c, rec
}

// TestAuthController_SignUpHandler のテストケース
func TestAuthController_SignUpHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		setupMock        func() *MockAuthService
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: 新規ユーザー登録成功",
			requestBody: `{"email":"test@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				userResponse := models.UserResponse{
					UserID: 1,
					Email:  "test@example.com",
					Role:   "customer",
				}
				token := "test-jwt-token"

				mockService.On("SignUp", mock.Anything, mock.MatchedBy(func(req models.AuthenticateRequest) bool {
					return req.Email == "test@example.com"
				})).Return(userResponse, token, nil)

				return mockService
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test-jwt-token", response.Token)
				assert.Equal(t, 1, response.User.UserID)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "customer", response.User.Role)
			},
		},
		{
			name:        "正常系: ゲストトークン付き新規ユーザー登録成功",
			requestBody: `{"email":"test@example.com","guest_order_token":"15ff4999-2cfd-41f3-b744-926e7c5c7a0e"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				userResponse := models.UserResponse{
					UserID: 1,
					Email:  "test@example.com",
					Role:   "customer",
				}
				token := "test-jwt-token"

				mockService.On("SignUp", mock.Anything, mock.MatchedBy(func(req models.AuthenticateRequest) bool {
					return req.Email == "test@example.com" && req.GuestOrderToken == "15ff4999-2cfd-41f3-b744-926e7c5c7a0e"
				})).Return(userResponse, token, nil)

				return mockService
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test-jwt-token", response.Token)
				assert.Equal(t, 1, response.User.UserID)
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"email":}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー（不正なメールアドレス）",
			requestBody: `{"email":"invalid-email"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 既に存在するユーザー",
			requestBody: `{"email":"existing@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				mockService.On("SignUp", mock.Anything, mock.Anything).Return(
					models.UserResponse{}, "", apperrors.Conflict.Wrap(nil, "指定されたメールアドレスは既に使用されています"))

				return mockService
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
			expectedCode:   apperrors.Conflict,
		},
		{
			name:        "異常系: サービス層で内部エラー",
			requestBody: `{"email":"test@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				mockService.On("SignUp", mock.Anything, mock.Anything).Return(
					models.UserResponse{}, "", apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))

				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAuthController(mockService)

			// テストコンテキストの作成
			c, rec := createTestContextForAuth(http.MethodPost, "/auth/signup", tt.requestBody)

			// テスト実行
			err := controller.SignUpHandler(c)

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

// TestAuthController_LogInHandler のテストケース
func TestAuthController_LogInHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		setupMock        func() *MockAuthService
		expectedStatus   int
		expectError      bool
		expectedCode     apperrors.ErrCode
		validateResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "正常系: ログイン成功",
			requestBody: `{"email":"test@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				userResponse := models.UserResponse{
					UserID: 1,
					Email:  "test@example.com",
					Role:   "customer",
				}
				token := "test-jwt-token"

				mockService.On("LogIn", mock.Anything, mock.MatchedBy(func(req models.AuthenticateRequest) bool {
					return req.Email == "test@example.com"
				})).Return(userResponse, token, nil)

				return mockService
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test-jwt-token", response.Token)
				assert.Equal(t, 1, response.User.UserID)
				assert.Equal(t, "test@example.com", response.User.Email)
				assert.Equal(t, "customer", response.User.Role)
			},
		},
		{
			name:        "正常系: ゲストトークン付きログイン成功",
			requestBody: `{"email":"test@example.com","guest_order_token":"15ff4999-2cfd-41f3-b744-926e7c5c7a0e"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				userResponse := models.UserResponse{
					UserID: 1,
					Email:  "test@example.com",
					Role:   "customer",
				}
				token := "test-jwt-token"

				mockService.On("LogIn", mock.Anything, mock.MatchedBy(func(req models.AuthenticateRequest) bool {
					return req.Email == "test@example.com" && req.GuestOrderToken == "15ff4999-2cfd-41f3-b744-926e7c5c7a0e"
				})).Return(userResponse, token, nil)

				return mockService
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test-jwt-token", response.Token)
				assert.Equal(t, 1, response.User.UserID)
			},
		},
		{
			name:        "正常系: 管理者ユーザーログイン成功",
			requestBody: `{"email":"admin@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				userResponse := models.UserResponse{
					UserID: 2,
					Email:  "admin@example.com",
					Role:   "admin",
				}
				token := "admin-jwt-token"

				mockService.On("LogIn", mock.Anything, mock.MatchedBy(func(req models.AuthenticateRequest) bool {
					return req.Email == "admin@example.com"
				})).Return(userResponse, token, nil)

				return mockService
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
			validateResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.AuthResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "admin-jwt-token", response.Token)
				assert.Equal(t, 2, response.User.UserID)
				assert.Equal(t, "admin@example.com", response.User.Email)
				assert.Equal(t, "admin", response.User.Role)
			},
		},
		{
			name:        "異常系: 不正なJSONリクエスト",
			requestBody: `{"email":}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ReqBodyDecodeFailed,
		},
		{
			name:        "異常系: バリデーションエラー（不正なメールアドレス）",
			requestBody: `{"email":"invalid-email"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)
				return mockService
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			expectedCode:   apperrors.ValidationFailed,
		},
		{
			name:        "異常系: 認証失敗（ユーザーが見つからない）",
			requestBody: `{"email":"notfound@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				mockService.On("LogIn", mock.Anything, mock.Anything).Return(
					models.UserResponse{}, "", apperrors.Unauthorized.Wrap(nil, "認証に失敗しました"))

				return mockService
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			expectedCode:   apperrors.Unauthorized,
		},
		{
			name:        "異常系: サービス層で内部エラー",
			requestBody: `{"email":"test@example.com"}`,
			setupMock: func() *MockAuthService {
				mockService := new(MockAuthService)

				mockService.On("LogIn", mock.Anything, mock.Anything).Return(
					models.UserResponse{}, "", apperrors.Unknown.Wrap(nil, "内部エラーが発生しました"))

				return mockService
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			expectedCode:   apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックのセットアップ
			mockService := tt.setupMock()
			defer mockService.AssertExpectations(t)

			// コントローラーの作成
			controller := controllers.NewAuthController(mockService)

			// テストコンテキストの作成
			c, rec := createTestContextForAuth(http.MethodPost, "/auth/login", tt.requestBody)

			// テスト実行
			err := controller.LogInHandler(c)

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
