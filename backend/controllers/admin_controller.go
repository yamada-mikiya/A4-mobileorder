package controllers

import (
	"net/http"
	"strconv"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
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
		return apperrors.BadParam.Wrap(err, "店舗IDの形式が不正です。")
	}

	claims, err := GetClaims(ctx)
	if err != nil {
		return err
	}
	if err := AuthorizeShopAccess(claims, targetShopID); err != nil {
		return err
	}

	orderList, err := c.s.GetCookingOrders(ctx.Request().Context(), targetShopID)
	if err != nil {
		return err
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
		return apperrors.BadParam.Wrap(err, "店舗IDの形式が不正です。")
	}

	claims, err := GetClaims(ctx)
	if err != nil {
		return err
	}
	if err := AuthorizeShopAccess(claims, targetShopID); err != nil {
		return err
	}

	orderList, err := c.s.GetCompletedOrders(ctx.Request().Context(), targetShopID)
	if err != nil {
		return err
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

	claims, err := GetClaims(ctx)
	if err != nil {
		return err
	}
	if claims.ShopID == nil {
		return apperrors.Forbidden.Wrap(nil, "店舗に紐づいていない管理者アカウントです。")
	}
	adminShopID := *claims.ShopID

	targetOrderID, err := strconv.Atoi(ctx.Param("order_id"))
	if err != nil {
		return apperrors.BadParam.Wrap(err, "注文IDの形式が不正です。")
	}

	err = c.s.UpdateOrderStatus(ctx.Request().Context(), adminShopID, targetOrderID)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "注文ステータスを更新しました。"})
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

	claims, err := GetClaims(ctx)
	if err != nil {
		return err
	}

	if claims.ShopID == nil {
		return apperrors.Forbidden.Wrap(nil, "店舗に紐づいていない管理者アカウントです。")
	}
	adminShopID := *claims.ShopID

	targetOrderID, err := strconv.Atoi(ctx.Param("order_id"))
	if err != nil {
		return apperrors.BadParam.Wrap(err, "注文IDの形式が不正です。")
	}

	err = c.s.DeleteOrder(ctx.Request().Context(), adminShopID, targetOrderID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]string{"message": "注文を削除しました。"})
}

// UpdateItemAvailabilityHandler は商品の販売状態を更新します
// @Summary      商品の販売状態を更新 (Admin)
// @Description  管理者が商品の販売可能状態を更新します（在庫切れ設定など）
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        item_id   path      int                              true  "商品ID"
// @Param        request   body      models.UpdateItemAvailabilityRequest true  "販売状態更新リクエスト"
// @Success      200       {object}  map[string]string                     "更新成功"
// @Failure      400       {object}  apperrors.ErrorResponse               "リクエストエラー"
// @Failure      404       {object}  apperrors.ErrorResponse               "商品が見つからない"
// @Failure      500       {object}  apperrors.ErrorResponse               "内部サーバーエラー"
// @Router       /admin/items/{item_id}/availability [patch]
// @Security     BearerAuth
func (c *adminController) UpdateItemAvailabilityHandler(ctx echo.Context) error {
	// パスパラメータから商品IDを取得
	itemIDParam := ctx.Param("item_id")
	itemID, err := strconv.Atoi(itemIDParam)
	if err != nil {
		return apperrors.BadParam.Wrap(err, "商品IDの形式が不正です。")
	}

	// リクエストボディをバインド
	var req models.UpdateItemAvailabilityRequest
	if err := ctx.Bind(&req); err != nil {
		return apperrors.BadParam.Wrap(err, "リクエスト形式が不正です。")
	}

	// バリデーション
	if err := ctx.Validate(req); err != nil {
		return apperrors.ValidationFailed.Wrap(err, "入力値が不正です。")
	}

	// サービス層で更新処理
	if err := c.s.UpdateItemAvailability(ctx.Request().Context(), itemID, req.IsAvailable); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "商品の販売状態を更新しました。",
	})
}
