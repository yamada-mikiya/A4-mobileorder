package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/golang-jwt/jwt/v5"
)

type AuthServicer interface {
	SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
	createToken(user models.User) (string, error)
	LogIn(ctx context.Context, req models.AuthenticateRequest) (string, error)
}

type authService struct {
	usr repositories.UserRepository
	shr repositories.ShopRepository
	orr repositories.OrderRepository
}

func NewAuthService(usr repositories.UserRepository, shr repositories.ShopRepository, orr repositories.OrderRepository) AuthServicer {
	return &authService{usr, shr, orr}
}

func (s *authService) SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	if !strings.Contains(req.Email, "@") {
		return models.UserResponse{}, "", errors.New("invalid email format")
	}

	newUser := &models.User{Email: req.Email}
	if err := s.usr.CreateUser(ctx, newUser); err != nil {
		// "email already exists" エラーをキャッチ
		return models.UserResponse{}, "", err
	}

	if req.UserOrderToken != "" {
		if err := s.orr.UpdateUserIDByGuestToken(ctx, req.UserOrderToken, newUser.UserID); err != nil {
			log.Printf("warning: failed to claim guest order for new user %d: %v", newUser.UserID, err)
		}
	}

	tokenString, err := s.createToken(*newUser)
	if err != nil {
		return models.UserResponse{}, "", fmt.Errorf("user created, but failed to create token: %w", err)
	}

	resUser := models.UserResponse{
		UserID: newUser.UserID,
		Email:  newUser.Email,
		Role:   newUser.Role.String(),
	}
	return resUser, tokenString, nil
}

func (s *authService) LogIn(ctx context.Context, req models.AuthenticateRequest) (string, error) {

	if !strings.Contains(req.Email, "@") {
		return "", errors.New("invalid email format")
	}

	storedUser, err := s.usr.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", err
	}

	if req.UserOrderToken != "" {
		if err := s.orr.UpdateUserIDByGuestToken(ctx, req.UserOrderToken, storedUser.UserID); err != nil {
			log.Printf("warning: failed to claim guest order for existing user %d: %v", storedUser.UserID, err)
		}
	}

	return s.createToken(storedUser)
}

func (s *authService) createToken(user models.User) (string, error) {
	claims := &models.JwtCustomClaims{
		UserID: user.UserID,
		Role:   user.Role,
	}

	if user.Role == models.AdminRole {
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
