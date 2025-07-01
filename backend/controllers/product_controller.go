package controllers

import (
	"net/http"
	"strconv"

	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type ProductController interface {
	GetProductListHandler(ctx echo.Context) error
}

type productController struct {
	s services.ProductServicer
}

func NewProductController(s services.ProductServicer) ProductController {
	return &productController{s}
}

// 商品一覧を取得
//いずみん
func (c *productController) GetProductListHandler(ctx echo.Context) error {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "invalid order_id")
	}

	productList, err := c.s.GetProductList(shopID)

	return ctx.JSON(http.StatusOK, productList)
}
