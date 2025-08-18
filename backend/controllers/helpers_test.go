package controllers_test

import (
	"testing"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)
func TestGetClaims(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() echo.Context
		expectError    bool
		expectedCode   apperrors.ErrCode
		expectMessage  string
		validateClaims func(t *testing.T, claims *models.JwtCustomClaims)
	}{
		{
			name: "正常系: 有効なJWTトークンからクレーム取得成功",
			setupContext: func() echo.Context {
				claims := &models.JwtCustomClaims{
					UserID: 123,
					Role:   models.AdminRole,
					ShopID: intPtr(456),
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

				e := echo.New()
				c := e.NewContext(nil, nil)
				c.Set("user", token)
				return c
			},
			expectError: false,
			validateClaims: func(t *testing.T, claims *models.JwtCustomClaims) {
				assert.Equal(t, 123, claims.UserID)
				assert.Equal(t, models.AdminRole, claims.Role)
				assert.Equal(t, 456, *claims.ShopID)
			},
		},
		{
			name: "異常系: トークンが設定されていない",
			setupContext: func() echo.Context {
				e := echo.New()
				return e.NewContext(nil, nil)
			},
			expectError:   true,
			expectedCode:  apperrors.Unauthorized,
			expectMessage: "リクエストにトークンが含まれていません",
		},
		{
			name: "異常系: トークンがnil",
			setupContext: func() echo.Context {
				e := echo.New()
				c := e.NewContext(nil, nil)
				c.Set("user", (*jwt.Token)(nil))
				return c
			},
			expectError:   true,
			expectedCode:  apperrors.Unauthorized,
			expectMessage: "リクエストにトークンが含まれていません",
		},
		{
			name: "異常系: 不正な型のトークン",
			setupContext: func() echo.Context {
				e := echo.New()
				c := e.NewContext(nil, nil)
				c.Set("user", "invalid-token-type")
				return c
			},
			expectError:   true,
			expectedCode:  apperrors.Unauthorized,
			expectMessage: "リクエストにトークンが含まれていません",
		},
		{
			name: "異常系: 不正なクレーム型",
			setupContext: func() echo.Context {
				token := &jwt.Token{
					Header: map[string]interface{}{
						"typ": "JWT",
						"alg": jwt.SigningMethodHS256.Alg(),
					},
					Claims: jwt.MapClaims{"user_id": "123"}, // 不正な型
					Method: jwt.SigningMethodHS256,
					Valid:  true,
				}

				e := echo.New()
				c := e.NewContext(nil, nil)
				c.Set("user", token)
				return c
			},
			expectError:   true,
			expectedCode:  apperrors.Unauthorized,
			expectMessage: "トークンクレームの解析に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ
			c := tt.setupContext()

			// テスト実行
			claims, err := controllers.GetClaims(c)

			// 結果検証
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
					assert.Contains(t, appErr.Error(), tt.expectMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				if tt.validateClaims != nil {
					tt.validateClaims(t, claims)
				}
			}
		})
	}
}

func TestAuthorizeShopAccess(t *testing.T) {
	tests := []struct {
		name          string
		claims        *models.JwtCustomClaims
		targetShopID  int
		expectError   bool
		expectedCode  apperrors.ErrCode
		expectMessage string
	}{
		{
			name: "正常系: 同じ店舗IDでアクセス権限あり",
			claims: &models.JwtCustomClaims{
				UserID: 123,
				Role:   models.AdminRole,
				ShopID: intPtr(456),
			},
			targetShopID: 456,
			expectError:  false,
		},
		{
			name: "異常系: 店舗IDがnilの管理者",
			claims: &models.JwtCustomClaims{
				UserID: 123,
				Role:   models.AdminRole,
				ShopID: nil, // 店舗に紐づいていない
			},
			targetShopID:  456,
			expectError:   true,
			expectedCode:  apperrors.Forbidden,
			expectMessage: "店舗に紐づいていない管理者アカウント",
		},
		{
			name: "異常系: 異なる店舗IDでアクセス権限なし",
			claims: &models.JwtCustomClaims{
				UserID: 123,
				Role:   models.AdminRole,
				ShopID: intPtr(456), // 店舗ID: 456
			},
			targetShopID:  789, // 異なる店舗ID
			expectError:   true,
			expectedCode:  apperrors.Forbidden,
			expectMessage: "この店舗へのアクセス権がありません",
		},
		{
			name: "境界値: 店舗ID=0の場合",
			claims: &models.JwtCustomClaims{
				UserID: 123,
				Role:   models.AdminRole,
				ShopID: intPtr(0),
			},
			targetShopID: 0,
			expectError:  false,
		},
		{
			name: "境界値: 大きな店舗IDの場合",
			claims: &models.JwtCustomClaims{
				UserID: 123,
				Role:   models.AdminRole,
				ShopID: intPtr(999999),
			},
			targetShopID: 999999,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト実行
			err := controllers.AuthorizeShopAccess(tt.claims, tt.targetShopID)

			// 結果検証
			if tt.expectError {
				assert.Error(t, err)
				if appErr, ok := err.(*apperrors.AppError); ok {
					assert.Equal(t, tt.expectedCode, appErr.ErrCode)
					assert.Contains(t, appErr.Error(), tt.expectMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// テスト用ヘルパー関数
func intPtr(i int) *int {
	return &i
}
