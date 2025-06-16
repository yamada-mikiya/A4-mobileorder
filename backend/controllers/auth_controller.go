package controllers

import (
	"net/http"

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
	return ctx.String(http.StatusOK, "Sign up")
}

func (c *AuthController) LogInHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Log in")
}
