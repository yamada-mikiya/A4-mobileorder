package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/A4-dev-team/mobileorder.git/validators"
	"github.com/labstack/echo/v4"
)

type OrderController interface {
	CreateAuthenticatedOrderHandler(ctx echo.Context) error
	CreateGuestOrderHandler(ctx echo.Context) error
	GetOrderListHandler(ctx echo.Context) error
	GetOrderStatusHandler(ctx echo.Context) error
}

type orderController struct {
	s services.OrderServicer
	v validators.Validator
}

func NewOrderController(s services.OrderServicer, v validators.Validator) OrderController {
	return &orderController{s, v}
}

// CreateAuthenticatedOrderHandler は認証済みユーザーの注文を作成します。
// @Summary      認証ユーザーの注文作成 (Create Order - Authenticated)
// @Description  認証済みのユーザーとして新しい注文を作成します。リクエストには有効なBearerトークンが必要です。
// @Tags         注文 (Order)
// @Accept       json
// @Produce      json
// @Param        shop_id path int true "店舗ID (Shop ID)"
// @Param        order body models.CreateOrderRequest true "注文内容 (Order details)"
// @Security     BearerAuth
// @Success      201 {object} models.AuthenticatedOrderResponse "作成された注文ID"
// @Failure      400 {object} map[string]string "リクエストボディまたは店舗IDが不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /shops/{shop_id}/orders [post]
func (c *orderController) CreateAuthenticatedOrderHandler(ctx echo.Context) error {
	reqItem := models.CreateOrderRequest{}
	if err := ctx.Bind(&reqItem); err != nil {
		return apperrors.ReqBodyDecodeFailed.Wrap(err, "リクエストの形式が不正です。")
	}

	if err := c.v.Validate(reqItem); err != nil {
		return apperrors.ValidationFailed.Wrap(err, err.Error())
	}

	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return apperrors.BadParam.Wrap(err, "店舗IDの形式が不正です。")
	}

	claims, err := getClaims(ctx)
	if err != nil {
		return err
	}
	userID := claims.UserID

	log.Printf("Authenticated user (ID: %d) order flow", userID)

	createdOrder, err := c.s.CreateAuthenticatedOrder(ctx.Request().Context(), userID, shopID, reqItem.Items)
	if err != nil {
		return err
	}

	resOrder := models.AuthenticatedOrderResponse{
		OrderID: uint(createdOrder.OrderID),
	}

	return ctx.JSON(http.StatusCreated, resOrder)
}

// CreateGuestOrderHandler はゲストユーザーの注文を作成します。
// @Summary      ゲストの注文作成 (Create Order - Guest)
// @Description  未ログインのゲストユーザーとして新しい注文を作成します。認証は不要です。
// @Tags         注文 (Order)
// @Accept       json
// @Produce      json
// @Param        shop_id path int true "店舗ID (Shop ID)"
// @Param        order body models.CreateOrderRequest true "注文内容 (Order details)"
// @Success      201 {object} models.CreateOrderResponse "作成された注文IDとゲスト用トークン"
// @Failure      400 {object} map[string]string "リクエストボディまたは店舗IDが不正です"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /shops/{shop_id}/guest-orders [post]
func (c *orderController) CreateGuestOrderHandler(ctx echo.Context) error {
	reqItem := models.CreateOrderRequest{}
	if err := ctx.Bind(&reqItem); err != nil {
		return apperrors.ReqBodyDecodeFailed.Wrap(err, "リクエストの形式が不正です。")
	}
	if err := c.v.Validate(reqItem); err != nil {
		return apperrors.ValidationFailed.Wrap(err, err.Error())
	}

	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return apperrors.BadParam.Wrap(err, "店舗IDの形式が不正です。")
	}

	log.Println("Guest user order flow")

	createdOrder, err := c.s.CreateOrder(ctx.Request().Context(), shopID, reqItem.Items)
	if err != nil {
		return err
	}

	resOrder := models.CreateOrderResponse{
		OrderID:         createdOrder.OrderID,
		GuestOrderToken: createdOrder.GuestOrderToken.String,
		Message:         "Order created successfully as a guest. Please sign up to claim this order.",
	}

	return ctx.JSON(http.StatusCreated, resOrder)
}

// GetOrderListHandler は、ユーザーが確認するユーザーのアクティブな注文履歴を取得します。
// @Summary      アクティブな注文履歴の取得 (Get Active Order List)
// @Description  ログイン中のユーザーの、現在アクティブな（調理中または調理完了）注文履歴を取得します。このAPIは常に'cooking'と'completed'ステータスの注文のみを返します。
// @Tags         注文 (Order)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} models.OrderListResponse "アクティブな注文履歴のリスト"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /orders [get]
func (c *orderController) GetOrderListHandler(ctx echo.Context) error {

	claims, err := getClaims(ctx)
	if err != nil {
		return err
	}
	userID := claims.UserID

	orderList, err := c.s.GetUserOrders(ctx.Request().Context(), userID)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, orderList)
}

// GetOrderStatusHandler は特定の注文ステータスを取得します。
// @Summary      注文ステータスの取得 (Get Order Status)
// @Description  特定の注文IDの現在のステータスと待ち状況をリアルタイムで取得します。
// @Tags         注文 (Order)
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id path int true "注文ID (Order ID)"
// @Success      200 {object} models.OrderStatusResponse "現在の注文ステータス"
// @Failure      400 {object} map[string]string "注文IDの形式が不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました"
// @Failure      404 {object} map[string]string "注文が見つからないか、アクセス権がありません"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /orders/{order_id}/status [get]
func (c *orderController) GetOrderStatusHandler(ctx echo.Context) error {
	orderIDStr := ctx.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return apperrors.BadParam.Wrap(err, "注文IDの形式が不正です。")
	}

	claims, err := getClaims(ctx)
	if err != nil {
		return err
	}
	userID := claims.UserID

	status, err := c.s.GetOrderStatus(ctx.Request().Context(), userID, orderID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, status)
}
