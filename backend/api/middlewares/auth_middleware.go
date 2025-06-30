package middlewares

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return next(c)
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, new(models.JwtCustomClaims), func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(os.Getenv("SECRET")), nil
		})

		if err != nil {
			log.Printf("JWT parsing error: %v", err)
		}

		if err == nil && token.Valid {
			c.Set("user", token)
			log.Println("JWT token is valid, user set to context.")
		} else {
			log.Println("JWT token is invalid or expired, proceeding as guest.") 
		}

		return next(c)
	}
}

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