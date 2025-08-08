package middlewares

import (
	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return apperrors.Unknown.Wrap(nil, "JWT token context is invalid")
		}
		claims, ok := user.Claims.(*models.JwtCustomClaims)
		if !ok {
			return apperrors.Unknown.Wrap(nil, "JWT claims are not of the expected type")
		}

		if claims.Role != models.AdminRole {
			return apperrors.Forbidden.Wrap(nil, "この操作を行う権限がありません。管理者アカウントが必要です。")
		}

		return next(c)
	}
}
