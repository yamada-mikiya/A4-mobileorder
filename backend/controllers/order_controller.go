package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type OrderController interface {
	GetProductListHandler(ctx echo.Context) error
	CreateOrderHandler(ctx echo.Context) error
	GetOrderListHandler(ctx echo.Context) error
	GetOrderStatusHandler(ctx echo.Context) error
}

type orderController struct {
	service services.OrderServicer
}

func NewOrderController(s services.OrderServicer) OrderController {
	return &orderController{service: s}
}

// 商品一覧を取得
func (c *orderController) GetProductListHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get product list")
}

// 新しい注文を作成
func (c *orderController) CreateOrderHandler(ctx echo.Context) error {
	// これで reqOrder.Products[0].ProductID のようにデータにアクセスできる
	reqProd := models.CreateOrderRequest{}
	if err := ctx.Bind(&reqProd); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	shopIDStr := ctx.Param("shop_id")

	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid shop ID format"})
	}

	userToken, ok := ctx.Get("user").(*jwt.Token)
	//jwt有効なときとjwtが有効でないときで処理を変える
	if ok && userToken != nil {
		log.Printf("Authenticated user order flow")
		claims := userToken.Claims.(*models.JwtCustomClaims)
		userID := claims.UserID

		createdOrder, err := c.service.CreateAuthenticatedOrder(ctx.Request().Context(), userID, shopID, reqProd.Products)
		if err != nil {
			log.Printf("Error creating authenticated order: %v", err)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create order"})
		}
		return ctx.JSON(http.StatusCreated, map[string]interface{}{"order_id": createdOrder.OrderID})
	} else {
		log.Println("Guest user order flow")
		createdOrder, err := c.service.CreateOrder(ctx.Request().Context(), shopID, reqProd.Products)
		if err != nil {
			log.Printf("Error creating guest order: %v", err)
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "fail to create order"})
		}

		resOrder := models.CreateOrderResponse{
			OrderID:        createdOrder.OrderID,
			UserOrderToken: createdOrder.UserOrderToken.String,
			Message:        "Order created successfully as a guest. Please sign up to claim this order.",
		}

		return ctx.JSON(http.StatusCreated, resOrder)

	}

}

// ユーザーの注文詳細を取得
func (c *orderController) GetOrderListHandler(ctx echo.Context) error {
	statusParams := ctx.QueryParams()["status"]
	if len(statusParams) == 0 {
		return ctx.JSON(http.StatusBadRequest, "at least one status query parameter is required")
	}

	userToken := ctx.Get("user").(jwt.Token)
	claims := userToken.Claims.(*models.JwtCustomClaims)
	userID := claims.UserID

	orderList, err := c.service.GetOrderList(ctx.Request().Context(), userID, statusParams)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "failed to get order list"})
	}

	return ctx.JSON(http.StatusOK, orderList)
}

// 注文のステータスを取得
func (c *orderController) GetOrderStatusHandler(ctx echo.Context) error {
	orderIDStr := ctx.Param("order_id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "invalid order_id")
	}

	userToken := ctx.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*models.JwtCustomClaims)
	userID := claims.UserID

	status, err := c.service.GetOrderStatus(ctx.Request().Context(), userID, orderID)
	if err != nil {
		if err.Error() == "order not found or you do not have permission" {
			return ctx.JSON(http.StatusNotFound, map[string]string{"message": err.Error()})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "failed to get order status"})
	}
	return ctx.JSON(http.StatusOK, status)
}
