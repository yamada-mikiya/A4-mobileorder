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
	//認証無し
	e.POST("/auth/signup", auc.SignUpHandler)
	e.POST("/auth/login", auc.LogInHandler)
	e.GET("/shops/:shop_id/products", orc.GetProductListHandler) // 商品一覧ページで商品情報取得
	e.POST("/shops/:shop_id/orders", orc.CreateOrderHandler, middlewares.OptionalAuth)                    // 注文作成

	//認証あり
	e.GET("/orders", orc.GetOrderDetailHandler, jwtMiddleware)        // ユーザーの注文確認（注文番号など）
	e.GET("/orders/:order_id/status", orc.GetOrderStatusHandler, jwtMiddleware) // 注文ステータスと待ち人数の取得(このエンドポイントを定期的に叩いてリアルタイムに近い更新を可能にする。)

	adminGroup := e.Group("/admin")
	adminGroup.Use(jwtMiddleware, middlewares.AdminRequired)
	{
		adminGroup.GET("/shops/:shop_id/orders", adc.GetAdminOrderListHandler)                       // 管理者の注文一覧（クエリで絞り込み）
		adminGroup.PATCH("/orders/:order_id/status", adc.UpdateOrderStatusHandler)                   // 管理者が注文ステータスを更新
		adminGroup.PATCH("/products/:product_id/availability", adc.UpdateProductAvailabilityHandler) // 商品の在庫状態更新
	}
	return e

}
