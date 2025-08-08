package services

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/golang-jwt/jwt/v5"
)

type AuthServicer interface {
	SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
	createToken(ctx context.Context, user models.User) (string, error)
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

	newUser := &models.User{Email: req.Email}
	if err := s.usr.CreateUser(ctx, newUser); err != nil {
		return models.UserResponse{}, "", err
	}

	if req.GuestOrderToken != "" {
		//GuestOrderTokenがあればユーザーと注文を結びつける
		if err := s.orr.UpdateUserIDByGuestToken(ctx, req.GuestOrderToken, newUser.UserID); err != nil {
			log.Printf("warning: failed to claim guest order for new user %d: %v", newUser.UserID, err)
		}
	}

	tokenString, err := s.createToken(ctx, *newUser)
	if err != nil {
		return models.UserResponse{}, "", err
	}

	resUser := models.UserResponse{
		UserID: newUser.UserID,
		Email:  newUser.Email,
		Role:   newUser.Role.String(),
	}
	return resUser, tokenString, nil
}

func (s *authService) LogIn(ctx context.Context, req models.AuthenticateRequest) (string, error) {

	storedUser, err := s.usr.GetUserByEmail(ctx, req.Email)
	if err != nil {
	var appErr *apperrors.AppError

    if errors.As(err, &appErr) {

        if appErr.ErrCode == apperrors.NoData {
            return "", apperrors.Unauthorized.Wrap(err, "メールアドレスが見つかりません。")
        }
    }
		if errors.As(err, &appErr) {

			if appErr.ErrCode == apperrors.NoData {
				return "", apperrors.Unauthorized.Wrap(err, "メールアドレスが見つかりません。")
			}
		}
		return "", apperrors.Unknown.Wrap(err, "ログイン処理中に予期せぬエラーが発生しました。")
	if errors.As(err, &appErr) {

		if appErr.ErrCode == apperrors.NoData {
			return "", apperrors.Unauthorized.Wrap(err, "メールアドレスが見つかりません。")
		}
	}
	return "", apperrors.Unknown.Wrap(err, "ログイン処理中に予期せぬエラーが発生しました。")
	if errors.As(err, &appErr) {

		if appErr.ErrCode == apperrors.NoData {
			return "", apperrors.Unauthorized.Wrap(err, "メールアドレスが見つかりません。")
		}
	}
	return "", apperrors.Unknown.Wrap(err, "ログイン処理中に予期せぬエラーが発生しました。")
	}

	if req.GuestOrderToken != "" {
		if err := s.orr.UpdateUserIDByGuestToken(ctx, req.GuestOrderToken, storedUser.UserID); err != nil {
			log.Printf("warning: failed to claim guest order for existing user %d: %v", storedUser.UserID, err)
		}
	}

	return s.createToken(ctx, storedUser)
}

func (s *authService) createToken(ctx context.Context, user models.User) (string, error) {
	claims := &models.JwtCustomClaims{
		UserID: user.UserID,
		Role:   user.Role,
	}

	if user.Role == models.AdminRole {
		shopID, err := s.shr.FindShopIDByAdminID(ctx, user.UserID)
		if err != nil {
			return "", apperrors.Unknown.Wrap(err, "管理者情報の取得に失敗し、トークンを生成できませんでした。")
		}
		claims.ShopID = &shopID
	}

	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		return "", apperrors.Unknown.Wrap(err, "トークンの署名に失敗しました。")
	}

	return tokenString, nil
}
