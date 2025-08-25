package services

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

var (
	secretKey     string
	secretKeyOnce sync.Once
)

func getSecretKey() (string, error) {
	secretKeyOnce.Do(func() {
		secretKey = os.Getenv("SECRET_KEY")
	})
	if secretKey == "" {
		return "", apperrors.Unknown.Wrap(errors.New("SECRET_KEY environment variable is not set or is empty"), "認証トークンの作成に失敗しました。")
	}
	return secretKey, nil
}

type AuthServicer interface {
	SignUp(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
	LogIn(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error)
}

type authService struct {
	usr repositories.UserRepository
	shr repositories.ShopRepository
	orr repositories.OrderRepository
	db  *sqlx.DB
}

func NewAuthService(usr repositories.UserRepository, shr repositories.ShopRepository, orr repositories.OrderRepository, db *sqlx.DB) AuthServicer {
	return &authService{
		usr: usr,
		shr: shr,
		orr: orr,
		db:  db,
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

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}
	defer tx.Rollback()

	// ユーザー作成（トランザクション内）
	newUser := &models.User{Email: req.Email}
	if err := s.usr.CreateUser(ctx, tx, newUser); err != nil {
		return models.UserResponse{}, "", err
	}

	// ゲスト注文引き継ぎ（必須処理）
	if err := s.orr.UpdateUserIDByGuestToken(ctx, tx, req.GuestOrderToken, newUser.UserID); err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "ゲスト注文の引き継ぎに失敗しました。")
	}

	// トークン生成
	token, err := s.createToken(ctx, *newUser)
	if err != nil {
		return models.UserResponse{}, "", err
	}
	tokenString = token

	userResponse = models.UserResponse{
		UserID: newUser.UserID,
		Email:  newUser.Email,
		Role:   newUser.Role.String(),
	}

	if err := tx.Commit(); err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
	}

	return userResponse, tokenString, nil
}

func (s *authService) signUpNormal(ctx context.Context, req models.AuthenticateRequest) (models.UserResponse, string, error) {
	newUser := &models.User{Email: req.Email}
	if err := s.usr.CreateUser(ctx, s.db, newUser); err != nil {
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
	storedUser, err := s.usr.GetUserByEmail(ctx, s.db, req.Email)
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

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}
	defer tx.Rollback()

	// ゲスト注文引き継ぎ（必須処理）
	if err := s.orr.UpdateUserIDByGuestToken(ctx, tx, req.GuestOrderToken, user.UserID); err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "ゲスト注文の引き継ぎに失敗しました。")
	}

	// トークン生成
	token, err := s.createToken(ctx, user)
	if err != nil {
		return models.UserResponse{}, "", err
	}
	tokenString = token

	userResponse = models.UserResponse{
		UserID: user.UserID,
		Email:  user.Email,
		Role:   user.Role.String(),
	}

	if err := tx.Commit(); err != nil {
		return models.UserResponse{}, "", apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
	}

	return userResponse, tokenString, nil
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
		shopID, err := s.shr.FindShopIDByAdminID(ctx, s.db, user.UserID)
		if err != nil {
			return "", apperrors.Unknown.Wrap(err, "店舗情報の取得に失敗しました。")
		}
		claims.ShopID = &shopID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key, err := getSecretKey()
	if err != nil {
		return "", err
	}

	t, err := token.SignedString([]byte(key))
	if err != nil {
		return "", apperrors.Unknown.Wrap(err, "認証トークンの作成に失敗しました。")
	}

	return t, nil
}
