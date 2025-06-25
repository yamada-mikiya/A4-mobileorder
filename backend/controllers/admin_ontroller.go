package controllers

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type AdminController interface {
	GetAdminOrderListHandler(ctx echo.Context) error
	UpdateOrderStatusHandler(ctx echo.Context) error
	UpdateProductAvailabilityHandler(ctx echo.Context) error
}

type adminController struct {
	service services.AdminServicer
}

func NewAdminController(s services.AdminServicer) AdminController {
	return &adminController{service: s}
}

func (c *adminController) GetAdminOrderListHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order list")
}

func (c *adminController) UpdateOrderStatusHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update order status ")
}

func (c *adminController) UpdateProductAvailabilityHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update product availability")
}
