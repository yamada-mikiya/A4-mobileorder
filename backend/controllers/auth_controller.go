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
// @Summary 新規ユーザー登録 (SignUp)
// @Tags Auth
// @Description 新しいユーザーアカウントを作成し、認証トークンとユーザー情報を返します。ゲスト注文トークンをリクエストに含めることで、既存のゲスト注文をアカウントに紐付けることも可能です。
// @Param body body models.AuthenticateRequest true "ユーザーのメールアドレスと、任意でゲスト注文トークン"
// @Accept json
// @Produce json
// @Success 201 {object} models.SignUpResponse "登録成功時のレスポンス"
// @Failure 400 {object} map[string]string "リクエストが不正な場合"
// @Failure 409 {object} map[string]string "メールアドレスが既に使用されている場合"
// @Failure 500 {object} map[string]string "サーバー内部エラー"
// @Router /auth/signup [post]
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
// @Summary ユーザー認証 (LogIn)
// @Tags Auth
// @Description 既存のユーザーを認証し、新しい認証トークンを発行します。ゲスト注文トークンをリクエストに含めることで、既存のゲスト注文をアカウントに紐付けることも可能です。
// @Param body body models.AuthenticateRequest true "ユーザーのメールアドレスと、任意でゲスト注文トークン"
// @Accept json
// @Produce json
// @Success 200 {object} models.LoginResponse "認証成功時のレスポンス"
// @Failure 400 {object} map[string]string "リクエストが不正な場合"
// @Failure 401 {object} map[string]string "認証に失敗した場合"
// @Failure 500 {object} map[string]string "サーバー内部エラー"
// @Router /auth/login [post]
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
