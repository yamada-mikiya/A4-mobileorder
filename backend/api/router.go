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

func NewRouter(adc controllers.IAdminController, auc controllers.IAuthController, orc controllers.IOrderController) *echo.Echo {
	e := echo.New()

	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(models.JwtCustomClaims)
		},
		SigningKey:    []byte(os.Getenv("SECRET")),
		TokenLookup:   "header:Authorization:Bearer ",
		SigningMethod: "HS256",
	}

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Authentication
	e.POST("/auth/signup", auc.SignUpHandler)
	e.POST("/auth/login", auc.LogInHandler)

	// for Customers
	e.GET("/shops/:shop_id/products", orc.GetProductListHandler) // 商品一覧取得
	customerGroup := e.Group("/orders")
	customerGroup.Use(echojwt.WithConfig(jwtConfig))

	customerGroup.POST("/orders", orc.CreateOrderHandler)                    // 注文作成
	customerGroup.GET("/orders/:order_id", orc.GetOrderDetailHandler)        // 注文確認（注文番号など）
	customerGroup.GET("/orders/:order_id/status", orc.GetOrderStatusHandler) // 注文ステータスと待ち人数の取得(このエンドポイントを定期的に叩いてリアルタイムに近い更新を可能にする。)

	// for Admin
	adminGroup := e.Group("/admin")
	adminGroup.Use(echojwt.WithConfig(jwtConfig))
	adminGroup.Use(middlewares.AdminRequired)

	adminGroup.GET("/admin/shops/:shop_id/orders", adc.GetAdminOrderListHandler)                       // 管理者の注文一覧（クエリで絞り込み）
	adminGroup.PATCH("/admin/orders/:order_id", adc.UpdateOrderStatusHandler)                          // 管理者が注文ステータスを更新
	adminGroup.PATCH("/admin/products/:product_id/availability", adc.UpdateProductAvailabilityHandler) // 商品の在庫状態更新

	return e

}
