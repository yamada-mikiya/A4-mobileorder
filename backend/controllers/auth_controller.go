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
	s services.AuthServicer
}

func NewAuthController(s services.AuthServicer) AuthController {
	return &authController{s}
}
// SignUpHandler は新しいユーザーアカウントを作成します。
// @Summary      新規ユーザー登録 (Sign Up)
// @Description  新しいユーザーアカウントを作成し、認証トークンとユーザー情報を返します。
// @Description  リクエストにゲスト注文トークンを含めることで、既存のゲスト注文をアカウントに紐付けることも可能です。
// @Tags         認証 (Auth)
// @Accept       json
// @Produce      json
// @Param        payload body models.AuthenticateRequest true "ユーザー情報 (メールアドレスと、任意でゲスト注文トークン)"
// @Success      201 {object} models.SignUpResponse "登録成功。ユーザー情報と認証トークンを返します。"
// @Failure      400 {object} map[string]string "リクエストボディが不正です"
// @Failure      409 {object} map[string]string "指定されたメールアドレスは既に使用されています"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /auth/signup [post]
func (c *authController) SignUpHandler(ctx echo.Context) error {
	req := models.AuthenticateRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request body"})
	}
	userRes, tokenString, err := c.s.SignUp(ctx.Request().Context(), req)
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
// LogInHandler は既存ユーザーを認証します。
// @Summary      ログイン (Log In)
// @Description  既存のユーザーを認証し、新しい認証トークンを発行します。
// @Description  リクエストにゲスト注文トークンを含めることで、既存のゲスト注文をアカウントに紐付けることも可能です。
// @Tags         認証 (Auth)
// @Accept       json
// @Produce      json
// @Param        payload body models.AuthenticateRequest true "ユーザー情報 (メールアドレスと、任意でゲスト注文トークン)"
// @Success      200 {object} models.LoginResponse "認証成功。新しい認証トークンを返します。"
// @Failure      400 {object} map[string]string "リクエストボディが不正です"
// @Failure      401 {object} map[string]string "認証に失敗しました (メールアドレスが存在しない等)"
// @Failure      500 {object} map[string]string "サーバー内部でエラーが発生しました"
// @Router       /auth/login [post]
func (c *authController) LogInHandler(ctx echo.Context) error {
	req := models.AuthenticateRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	tokenString, err := c.s.LogIn(ctx.Request().Context(), req)
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
