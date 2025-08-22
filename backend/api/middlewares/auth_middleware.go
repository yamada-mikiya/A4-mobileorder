package middlewares

import (
	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/labstack/echo/v4"
)

func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := controllers.GetClaims(c)
		if err != nil {
			return err
		}

		if claims.Role != models.AdminRole {
			return apperrors.Forbidden.Wrap(nil, "この操作を行う権限がありません。管理者アカウントが必要です。")
		}

		return next(c)
	}
}
