package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AdminController interface {
	GetCookingOrdersHandler(ctx echo.Context) error
	GetCompletedOrdersHandler(ctx echo.Context) error
	UpdateOrderStatusHandler(ctx echo.Context) error
	UpdateItemAvailabilityHandler(ctx echo.Context) error
	DeleteOrderHandler(ctx echo.Context) error
}

type adminController struct {
	s services.AdminServicer
}

func NewAdminController(s services.AdminServicer) AdminController {
	return &adminController{s}
}

// GetCookingOrdersHandler は、「調理中」の注文一覧を取得します。
// @Summary      「調理中」の注文一覧を取得 (Admin)
// @Description  ログイン中の管理者が担当する店舗の、「調理中」ステータスの注文を全て取得します。
// @Tags         管理者 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        shop_id path int true "注文一覧を取得する店舗のID"
// @Success      200 {array} models.AdminOrderResponse "調理中の注文リスト"
// @Failure      400 {object} map[string]string "店舗IDの形式が不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      403 {object} map[string]string "この店舗へのアクセス権がありません"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /admin/shops/{shop_id}/orders/cooking [get]
func (c *adminController) GetCookingOrdersHandler(ctx echo.Context) error {

	targetShopIDStr := ctx.Param("shop_id")
	targetShopID, err := strconv.Atoi(targetShopIDStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid shop_id format in URL."})
	}
	userToken, ok := ctx.Get("user").(*jwt.Token)
	if !ok || userToken == nil {
		return ctx.JSON(http.StatusUnauthorized, "Invalid or missing token.")
	}
	claims, ok := userToken.Claims.(*models.JwtCustomClaims)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, "Failed to parse custom claims.")
	}

	if claims.ShopID == nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You are an admin but not associated with any shop."})
	}
	adminShopID := *claims.ShopID

	if adminShopID != targetShopID {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You do not have permission to access this shop's data."})
	}

	log.Printf("Admin user (ID: %d) authorized to access orders for shop (ID: %d)", claims.UserID, targetShopID)

	orderList, err := c.s.GetCookingOrders(ctx.Request().Context(), targetShopID)
	if err != nil {
		log.Printf("failed to get cooking orders: %v", err)
		return ctx.JSON(http.StatusInternalServerError, "failed to get orders")
	}
	return ctx.JSON(http.StatusOK, orderList)
}

// GetCompletedOrdersHandler は、「調理完了」の注文一覧を取得します。
// @Summary      「調理完了」の注文一覧を取得 (Admin)
// @Description  ログイン中の管理者が担当する店舗の、「調理完了」ステータスの注文を全て取得します。
// @Tags         管理者 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        shop_id path int true "注文一覧を取得する店舗のID"
// @Success      200 {array} models.AdminOrderResponse "調理完了の注文リスト"
// @Failure      400 {object} map[string]string "店舗IDの形式が不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      403 {object} map[string]string "この店舗へのアクセス権がありません"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /admin/shops/{shop_id}/orders/completed [get]
func (c *adminController) GetCompletedOrdersHandler(ctx echo.Context) error {

	targetShopIDStr := ctx.Param("shop_id")
	targetShopID, err := strconv.Atoi(targetShopIDStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid shop_id format in URL."})
	}
	userToken, ok := ctx.Get("user").(*jwt.Token)
	if !ok || userToken == nil {
		return ctx.JSON(http.StatusUnauthorized, "Invalid or missing token.")
	}
	claims, ok := userToken.Claims.(*models.JwtCustomClaims)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, "Failed to parse custom claims.")
	}

	if claims.ShopID == nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You are an admin but not associated with any shop."})
	}
	adminShopID := *claims.ShopID

	if adminShopID != targetShopID {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You do not have permission to access this shop's data."})
	}

	log.Printf("Admin user (ID: %d) authorized to access orders for shop (ID: %d)", claims.UserID, targetShopID)

	orderList, err := c.s.GetCompletedOrders(ctx.Request().Context(), targetShopID)
	if err != nil {
		log.Printf("failed to get completed orders: %v", err)
		return ctx.JSON(http.StatusInternalServerError, "failed to get orders")
	}
	return ctx.JSON(http.StatusOK, orderList)
}

// UpdateOrderStatusHandler は、注文のステータスを一段階進めます。
// @Summary      注文ステータスの更新 (Admin)
// @Description  管理者が担当する店舗の注文ステータスを一段階進めます (調理中→調理完了→お渡し済み)。リクエストボディは不要です。
// @Tags         管理者 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id path int true "ステータスを更新する注文のID"
// @Success      200 {object} map[string]string "成功メッセージ"
// @Failure      400 {object} map[string]string "注文IDの形式が不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      403 {object} map[string]string "この注文へのアクセス権がありません"
// @Failure      404 {object} map[string]string "指定された注文が見つかりません"
// @Failure      409 {object} map[string]string "これ以上ステータスを進められない場合に返されます (例: 'handed'の注文)"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /admin/orders/{order_id}/status [patch]
func (c *adminController) UpdateOrderStatusHandler(ctx echo.Context) error {

	userToken, ok := ctx.Get("user").(*jwt.Token)
	if !ok || userToken == nil {
		return ctx.JSON(http.StatusUnauthorized, "Invalid or missing token.")
	}
	claims, ok := userToken.Claims.(*models.JwtCustomClaims)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, "Failed to parse custom claims.")
	}
	if claims.ShopID == nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You are not associated with any shop."})
	}
	adminShopID := *claims.ShopID

	targetOrderID, err := strconv.Atoi(ctx.Param("order_id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid order_id format."})
	}

	err = c.s.UpdateOrderStatus(ctx.Request().Context(), adminShopID, targetOrderID)
	if err != nil {

		if err.Error() == "order not found or permission denied" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		if strings.Contains(err.Error(), "cannot be advanced") {
			return ctx.JSON(http.StatusConflict, map[string]string{"message": err.Error()})
		}
		log.Printf("Failed to advance order status: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "An internal error occurred."})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Order status advanced successfully."})
}

// DeleteOrderHandler は、管理者が担当する店舗の注文を削除します。
// @Summary      注文の削除 (Admin)
// @Description  管理者が担当する店舗の注文を削除します。
// @Tags         管理者 (Admin)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id path int true "削除する注文のID"
// @Success      200 {object} map[string]string "成功メッセージ"
// @Failure      400 {object} map[string]string "注文IDの形式が不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      403 {object} map[string]string "この注文へのアクセス権がありません"
// @Failure      404 {object} map[string]string "指定された注文が見つかりません"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /admin/orders/{order_id}/delete [delete]
func (c *adminController) DeleteOrderHandler(ctx echo.Context) error {

	userToken, ok := ctx.Get("user").(*jwt.Token)
	if !ok || userToken == nil {
		return ctx.JSON(http.StatusUnauthorized, "Invalid or missing token.")
	}
	claims, ok := userToken.Claims.(*models.JwtCustomClaims)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, "Failed to parse custom claims.")
	}
	if claims.ShopID == nil {
		return ctx.JSON(http.StatusForbidden, map[string]string{"message": "You are not associated with any shop."})
	}
	adminShopID := *claims.ShopID

	targetOrderID, err := strconv.Atoi(ctx.Param("order_id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid order_id format."})
	}

	err = c.s.DeleteOrder(ctx.Request().Context(), adminShopID, targetOrderID)
	if err != nil {
		if err.Error() == "order not found or permission denied" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		log.Printf("Failed to delete order: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "An internal error occurred."})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Order deleted successfully."})
}

func (c *adminController) UpdateItemAvailabilityHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update item availability")
}
