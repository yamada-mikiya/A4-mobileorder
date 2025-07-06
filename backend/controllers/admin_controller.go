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
	GetAdminOrderListHandler(ctx echo.Context) error
	UpdateOrderStatusHandler(ctx echo.Context) error
	UpdateProductAvailabilityHandler(ctx echo.Context) error
	DeleteOrderHandler(ctx echo.Context) error
}

type adminController struct {
	s services.AdminServicer
}

func NewAdminController(s services.AdminServicer) AdminController {
	return &adminController{s}
}

func (c *adminController) GetAdminOrderListHandler(ctx echo.Context) error {

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

	pageData, err := c.s.GetAdminOrderPageData(ctx.Request().Context(), targetShopID)
	if err != nil {
		log.Printf("failed to get admin order page data: %v", err)
		return ctx.JSON(http.StatusInternalServerError, "failed to get orders")
	}
	return ctx.JSON(http.StatusOK, pageData)
}

// UpdateOrderStatusHandler は、注文のステータスを一段階進めます。
// PATCH /admin/shops/:shop_id/orders/:order_id/status のようなURLを想定
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

	err = c.s.AdvanceOrderStatus(ctx.Request().Context(), adminShopID, targetOrderID)
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

func (c *adminController) UpdateProductAvailabilityHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update product availability")
}
