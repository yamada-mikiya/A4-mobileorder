package services

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

type AuthServicer interface {
	SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
	LogIn(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
}

type authService struct {
	usr repositories.UserRepository
	shr repositories.ShopRepository
	orr repositories.OrderRepository
	txm TransactionManager
}

func NewAuthService(usr repositories.UserRepository, shr repositories.ShopRepository, orr repositories.OrderRepository, db *sqlx.DB) AuthServicer {
	return &authService{
		usr: usr,
		shr: shr,
		orr: orr,
		txm: NewTransactionManager(db),
	}
}

// NewAuthServiceForTest creates an auth service for unit testing with mocked dependencies
func NewAuthServiceForTest(usr repositories.UserRepository, shr repositories.ShopRepository, orr repositories.OrderRepository, txm TransactionManager) AuthServicer {
	return &authService{
		usr: usr,
		shr: shr,
		orr: orr,
		txm: txm,
	}
}

func (s *authService) SignUp(ctx context.Context, req models.AuthenticateRequest) (userResponse models.UserResponse, tokenString string, err error) {
	// ゲスト注文トークンがある場合はトランザクション化
	if req.GuestOrderToken != "" {
		return s.signUpWithGuestOrderLinking(ctx, req)
	}

	// 通常のサインアップ
	return s.signUpNormal(ctx, req)
}

func (s *authService) signUpWithGuestOrderLinking(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	var userResponse models.UserResponse
	var tokenString string

	err := s.txm.WithUserOrderTransaction(ctx, func(txUserRepo repositories.UserRepository, txOrderRepo repositories.OrderRepository) error {
		// ユーザー作成（トランザクション内）
		newUser := &models.User{Email: req.Email}
		if err := txUserRepo.CreateUser(ctx, newUser); err != nil {
			return err
		}

		// ゲスト注文引き継ぎ（必須処理）
		if err := txOrderRepo.UpdateUserIDByGuestToken(ctx, req.GuestOrderToken, newUser.UserID); err != nil {
			return apperrors.Unknown.Wrap(err, "ゲスト注文の引き継ぎに失敗しました。")
		}

		// トークン生成
		token, err := s.createToken(ctx, *newUser)
		if err != nil {
			return err
		}
		tokenString = token

		userResponse = models.UserResponse{
			UserID: newUser.UserID,
			Email:  newUser.Email,
			Role:   newUser.Role.String(),
		}
		return nil
	})

	return userResponse, tokenString, err
}

func (s *authService) signUpNormal(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	newUser := &models.User{Email: req.Email}
	if err := s.usr.CreateUser(ctx, newUser); err != nil {
		return models.UserResponse{}, "", err
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

func (s *authService) LogIn(ctx context.Context, req models.AuthenticateRequest) (userResponse models.UserResponse, tokenString string, err error) {
	// ユーザー情報取得
	storedUser, err := s.usr.GetUserByEmail(ctx, req.Email)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			if appErr.ErrCode == apperrors.NoData {
				return models.UserResponse{}, "", apperrors.Unauthorized.Wrap(err, "メールアドレスが見つかりません。")
			}
		}
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "ログイン処理中に予期せぬエラーが発生しました。")
	}

	// ゲスト注文トークンがある場合はトランザクション化
	if req.GuestOrderToken != "" {
		return s.logInWithGuestOrderLinking(ctx, req, storedUser)
	}

	// 通常のログイン
	return s.logInNormal(ctx, storedUser)
}

func (s *authService) logInWithGuestOrderLinking(ctx context.Context, req models.AuthenticateRequest, user models.User) (models.UserResponse, string, error) {
	var userResponse models.UserResponse
	var tokenString string

	err := s.txm.WithOrderTransaction(ctx, func(txOrderRepo repositories.OrderRepository) error {
		// ゲスト注文引き継ぎ（必須処理）
		if err := txOrderRepo.UpdateUserIDByGuestToken(ctx, req.GuestOrderToken, user.UserID); err != nil {
			return apperrors.Unknown.Wrap(err, "ゲスト注文の引き継ぎに失敗しました。")
		}

		// トークン生成
		token, err := s.createToken(ctx, user)
		if err != nil {
			return err
		}
		tokenString = token

		userResponse = models.UserResponse{
			UserID: user.UserID,
			Email:  user.Email,
			Role:   user.Role.String(),
		}
		return nil
	})

	return userResponse, tokenString, err
}

func (s *authService) logInNormal(ctx context.Context, user models.User) (models.UserResponse, string, error) {
	tokenString, err := s.createToken(ctx, user)
	if err != nil {
		return models.UserResponse{}, "", err
	}

	resUser := models.UserResponse{
		UserID: user.UserID,
		Email:  user.Email,
		Role:   user.Role.String(),
	}
	return resUser, tokenString, nil
}

func (s *authService) createToken(ctx context.Context, user models.User) (string, error) {
	claims := &models.JwtCustomClaims{
		UserID: user.UserID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // 72時間後に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if user.Role == models.AdminRole {
		shopID, err := s.shr.FindShopIDByAdminID(ctx, user.UserID)
		if err != nil {
			return "", apperrors.Unknown.Wrap(err, "店舗情報の取得に失敗しました。")
		}
		claims.ShopID = &shopID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return "", apperrors.Unknown.Wrap(errors.New("SECRET_KEY environment variable is not set or is empty"), "認証トークンの作成に失敗しました。")
	}

	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", apperrors.Unknown.Wrap(err, "認証トークンの作成に失敗しました。")
	}

	return t, nil
}
