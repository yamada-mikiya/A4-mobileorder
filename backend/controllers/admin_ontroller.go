package controllers

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type IAdminController interface {
	GetAdminOrderListHandler(ctx echo.Context) error
	UpdateOrderStatusHandler(ctx echo.Context) error
	UpdateProductAvailabilityHandler(ctx echo.Context) error
}

type AdminController struct {
	service services.IAdminServicer
}

func NewAdminController(s services.IAdminServicer) IAdminController {
	return &AdminController{service: s}
}

func (c *AdminController) GetAdminOrderListHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order list")
}

func (c *AdminController) UpdateOrderStatusHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update order status ")
}

func (c *AdminController) UpdateProductAvailabilityHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Update Product Availability")
}
