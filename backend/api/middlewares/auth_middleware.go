package middlewares

import (
	"net/http"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return c.JSON(http.StatusInternalServerError, "JWT token not found or invalid")
		}
		claims, ok := user.Claims.(*models.JwtCustomClaims)
		if !ok {
			return c.JSON(http.StatusInternalServerError, "Invalid cstom claims")
		}

		if claims.Role != models.AdminRole {
			return c.JSON(http.StatusForbidden, "Admins only: You do not have permission to access this resource")
		}

		return next(c)
	}
}
