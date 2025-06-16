package api

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/labstack/echo/v4"
)

func NewRouter(adc controllers.IAdminController, auc controllers.IAuthController, orc controllers.IOrderController) *echo.Echo {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Authentication
	e.POST("/auth/signup", auc.SignUpHandler)
	e.POST("/auth/login", auc.LogInHandler)

	// for Customers
	e.GET("/shops/:shop_id/products", orc.GetProductListHandler) // 商品一覧取得
	e.POST("/orders", orc.CreateOrderHandler)                    // 注文作成
	e.GET("/orders/:order_id", orc.GetOrderDetailHandler)        // 注文確認（注文番号など）
	e.GET("/orders/:order_id/status", orc.GetOrderStatusHandler) // 注文ステータスと待ち人数の取得(このエンドポイントを定期的に叩いてリアルタイムに近い更新を可能にする。)

	// for Admin
	e.GET("/admin/shops/:shop_id/orders", adc.GetAdminOrderListHandler)                       // 管理者の注文一覧（クエリで絞り込み）
	e.PATCH("/admin/orders/:order_id", adc.UpdateOrderStatusHandler)                          // 管理者が注文ステータスを更新
	e.PATCH("/admin/products/:product_id/availability", adc.UpdateProductAvailabilityHandler) // 商品の在庫状態更新

	return e

}
