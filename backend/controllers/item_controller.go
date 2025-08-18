package controllers

import (
	"net/http"
	"strconv"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type ItemController interface {
	GetItemListHandler(ctx echo.Context) error
}

type itemController struct {
	s services.ItemServicer
}

func NewItemController(s services.ItemServicer) ItemController {
	return &itemController{s}
}

// 商品一覧を取得
// いずみん
func (c *itemController) GetItemListHandler(ctx echo.Context) error {
	shopIDStr := ctx.Param("shop_id")
	shopID, err := strconv.Atoi(shopIDStr)
	if err != nil {
		return apperrors.BadParam.Wrap(err, "店舗IDの形式が不正です。")
	}

	itemList, err := c.s.GetItemList(shopID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, itemList)
}
