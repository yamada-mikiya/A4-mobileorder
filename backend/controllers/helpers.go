package controllers

import (
	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// GetClaims はEchoコンテキストからJWTクレームを取得します
func GetClaims(ctx echo.Context) (*models.JwtCustomClaims, error) {
	userToken, ok := ctx.Get("user").(*jwt.Token)
	if !ok || userToken == nil {
		return nil, apperrors.Unauthorized.Wrap(nil, "リクエストにトークンが含まれていません。")
	}
	claims, ok := userToken.Claims.(*models.JwtCustomClaims)
	if !ok {
		return nil, apperrors.Unauthorized.Wrap(nil, "トークンクレームの解析に失敗しました。")
	}
	return claims, nil
}

// AuthorizeShopAccess は管理者の店舗アクセス権限をチェックします
func AuthorizeShopAccess(claims *models.JwtCustomClaims, targetShopID int) error {
	if claims.ShopID == nil {
		return apperrors.Forbidden.Wrap(nil, "店舗に紐づいていない管理者アカウントです。")
	}
	if *claims.ShopID != targetShopID {
		return apperrors.Forbidden.Wrap(nil, "この店舗へのアクセス権がありません。")
	}
	return nil
}
