package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/golang-jwt/jwt/v5"
)

type AuthServicer interface {
	SignUp(user models.User) (models.UserResponse, string, error)
	createToken(user models.User) (string, error)
	LogIn(user models.User) (string, error)
}

type authService struct {
	usr repositories.UserRepository
	shr repositories.ShopRepository
}

func NewAuthService(usr repositories.UserRepository, shr repositories.ShopRepository) AuthServicer {
	return &authService{usr, shr}
}

func (s *authService) SignUp(user models.User) (models.UserResponse, string, error) {

	if !strings.Contains(user.Email, "@") {
		return models.UserResponse{}, "", errors.New("invalid email format")
	}

	newUser := models.User{Email: user.Email}
	if err := s.usr.CreateUser(&newUser); err != nil {
		return models.UserResponse{}, "", err
	}
	tokenString, err := s.createToken(newUser)
	if err != nil {
		return models.UserResponse{}, "", fmt.Errorf("user created, but failed to create token: %w", err)
	}

	resUser := models.UserResponse{
		UserID: newUser.UserID,
		Email:  newUser.Email,
	}
	return resUser, tokenString, nil
}

func (s *authService) createToken(user models.User) (string, error) {
	claims := &models.JwtCustomClaims{
		UserID: user.UserID,
		Role:   user.Role,
	}

	if user.Role == "admin" {
		shop, err := s.shr.GetShopByAdminID(user.UserID)
		if err != nil {
			return "", fmt.Errorf("admin user found but failed to get shop info: %w", err)
		}
		shopID := int(shop.ShopID)
		claims.ShopID = &shopID
	}

	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *authService) LogIn(user models.User) (string, error) {

	if !strings.Contains(user.Email, "@") {
		return "", errors.New("invalid email format")
	}

	storedUser, err := s.usr.GetUserByEmail(user.Email)
	if err != nil {
		return "", err
	}

	return s.createToken(storedUser)
}
