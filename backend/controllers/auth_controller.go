package controllers

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type IAuthController interface {
	SignUpHandler(ctx echo.Context) error
	LogInHandler(ctx echo.Context) error
}

type AuthController struct {
	service services.IAuthServicer
}

func NewAuthController(s services.IAuthServicer) IAuthController {
	return &AuthController{service: s}
}

func (c *AuthController) SignUpHandler(ctx echo.Context) error {
	user := models.User{}
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	userRes, tokenString, err := c.service.SignUp(user)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusCreated, map[string]interface{}{
		"user":  userRes,
		"token": tokenString,
	})
}

func (c *AuthController) LogInHandler(ctx echo.Context) error {
	user := models.User{}
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	tokenString, err := c.service.LogIn(user)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"token": tokenString,
	})
}
