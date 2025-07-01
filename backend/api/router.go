// /api/router.go

package api

import (
	"net/http"
	"os"

	"github.com/A4-dev-team/mobileorder.git/api/middlewares"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func NewRouter(adc controllers.AdminController, auc controllers.AuthController, orc controllers.OrderController) *echo.Echo {
	e := echo.New()

	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(models.JwtCustomClaims)
		},
		SigningKey:    []byte(os.Getenv("SECRET")),
		TokenLookup:   "header:Authorization:Bearer ",
		SigningMethod: "HS256",
	}
	jwtMiddleware := echojwt.WithConfig(jwtConfig)

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// --- 認証不要なエンドポイント ---
	e.POST("/auth/signup", auc.SignUpHandler)
	e.POST("/auth/login", auc.LogInHandler)
	e.GET("/shops/:shop_id/products", orc.GetProductListHandler)
	e.POST("/shops/:shop_id/guest-orders", orc.CreateGuestOrderHandler)

	// --- 認証が必要なエンドポイント ---
	e.POST("/shops/:shop_id/orders", orc.CreateAuthenticatedOrderHandler, jwtMiddleware)
	e.GET("/orders", orc.GetOrderListHandler, jwtMiddleware)
	e.GET("/orders/:order_id/status", orc.GetOrderStatusHandler, jwtMiddleware)

	// --- 管理者用エンドポイント（変更なし） ---
	adminGroup := e.Group("/admin")
	adminGroup.Use(jwtMiddleware, middlewares.AdminRequired)
	{
		adminGroup.GET("/shops/:shop_id/orders", adc.GetAdminOrderListHandler)                       // 管理者の注文一覧（クエリで絞り込み）
		adminGroup.PATCH("/orders/:order_id/status", adc.UpdateOrderStatusHandler)                   // 管理者が注文ステータスを更新
		adminGroup.PATCH("/products/:product_id/availability", adc.UpdateProductAvailabilityHandler) // 商品の在庫状態更新
	}
	return e
}
