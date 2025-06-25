package controllers

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type OrderController interface {
	GetProductListHandler(ctx echo.Context) error
	CreateOrderHandler(ctx echo.Context) error
	GetOrderDetailHandler(ctx echo.Context) error
	GetOrderStatusHandler(ctx echo.Context) error
}

type orderController struct {
	service services.OrderServicer
}

func NewOrderController(s services.OrderServicer) OrderController {
	return &orderController{service: s}
}

func (c *orderController) GetProductListHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get prooduct list")
}

func (c *orderController) CreateOrderHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Create order")
}

func (c *orderController) GetOrderDetailHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order detail")
}

func (c *orderController) GetOrderStatusHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order status")
}
