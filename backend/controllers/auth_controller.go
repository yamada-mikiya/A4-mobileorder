package controllers

import (
	"log"
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/services"
	"github.com/labstack/echo/v4"
)

type AuthController interface {
	SignUpHandler(ctx echo.Context) error
	LogInHandler(ctx echo.Context) error
}

type authController struct {
	service services.AuthServicer
}

func NewAuthController(s services.AuthServicer) AuthController {
	return &authController{service: s}
}
//@Summary signup handler
//@Tags Auth
//@Description Get token and user info from header
//@Param body body models.AuthenticateRequest true "use e-mail and token"
//@Accept json
//@Produce json
//@success 201 {object} models.SignUpResponse
//@Security BearerAuth
//@Router /auth/signup [post]
func (c *authController) SignUpHandler(ctx echo.Context) error {
	req := models.AuthenticateRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}
	userRes, tokenString, err := c.service.SignUp(ctx.Request().Context(), req)
	if err != nil {
		if err.Error() == "email already exists" {
			return ctx.JSON(http.StatusConflict, map[string]string{"message": "このメールアドレスは既に使用されています。"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "サインアップ処理中にエラーが発生しました。"})
	}
	return ctx.JSON(http.StatusCreated, map[string]interface{}{
		"user":  userRes,
		"token": tokenString,
	})
}
//@Summary login handler
//@Tags Auth
//@Description Get token from header and Userid
//@Param body body models.AuthenticateRequest true "use e-mail and token"
//@Accept json
//@Produce json
//@success 201 {object} models.LoginResponse
//@Security BearerAuth
//@Router /auth/login [post]
func (c *authController) LogInHandler(ctx echo.Context) error {
	req := models.AuthenticateRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	tokenString, err := c.service.LogIn(ctx.Request().Context(), req)
	if err != nil {
		if err.Error() == "user not found" {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{"message": "メールアドレスが見つかりません。"})
		}
		log.Printf("Internal error during LogIn service: %v", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"message": "ログイン処理中にエラーが発生しました。"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"token": tokenString,
	})
}
