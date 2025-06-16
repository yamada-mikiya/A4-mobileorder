package controllers

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type IOrderController interface {
	GetProductListHandler(ctx echo.Context) error
	CreateOrderHandler(ctx echo.Context) error
	GetOrderDetailHandler(ctx echo.Context) error
	GetOrderStatusHandler(ctx echo.Context) error
}

type OrderController struct {
	service services.IOrderServicer
}

func NewOrderController(s services.IOrderServicer) IOrderController {
	return &OrderController{service: s}
}

func (c *OrderController) GetProductListHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get prooduct list")
}

func (c *OrderController) CreateOrderHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Create order")
}

func (c *OrderController) GetOrderDetailHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order detail")
}

func (c *OrderController) GetOrderStatusHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Get order status")
}
