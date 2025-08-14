package repositories_test

import (
	"context"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/models"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

// テスト定数
const (
	testEmail1       = "test1@example.com"
	testEmail2       = "test2@example.com"
	duplicateEmail   = "duplicate@example.com"
	nonExistentEmail = "nonexistent@example.com"
)

// createTestUserWithEmail - 指定されたメールアドレスでテスト用ユーザーを作成するヘルパー関数
func createTestUserWithEmail(t *testing.T, tx *sqlx.Tx, email string) *models.User {
	t.Helper()

	user := &models.User{
		Email: email,
		Role:  models.CustomerRole,
	}

	query := `INSERT INTO users (email, role) VALUES ($1, $2) RETURNING user_id, created_at, updated_at`
	err := tx.QueryRowx(query, user.Email, user.Role).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("テスト用ユーザーの作成に失敗しました: %v", err)
	}

	return user
}

// TestCreateUser - ユーザー作成のテスト
func TestCreateUser(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		user            *models.User
		setup           func(*sqlx.Tx)
		expectedErrCode apperrors.ErrCode
		assertion       func(*testing.T, *models.User)
	}{
		{
			name: "正常系: 新規ユーザーが正常に作成される",
			user: &models.User{
				Email: testEmail1,
			},
			setup: func(tx *sqlx.Tx) {},
			assertion: func(t *testing.T, user *models.User) {
				// 作成されたユーザーの検証
				if user.UserID == 0 {
					t.Error("UserID が設定されていません")
				}
				if user.Role != models.CustomerRole {
					t.Errorf("Role: expected %s, got %s", models.CustomerRole, user.Role)
				}
				if user.CreatedAt.IsZero() {
					t.Error("CreatedAt が設定されていません")
				}
				if user.UpdatedAt.IsZero() {
					t.Error("UpdatedAt が設定されていません")
				}
			},
		},
		{
			name: "異常系: 重複するメールアドレスの場合はConflictエラーを返す",
			user: &models.User{
				Email: duplicateEmail,
			},
			setup: func(tx *sqlx.Tx) {
				// 事前に同じメールアドレスのユーザーを作成
				createTestUserWithEmail(t, tx, duplicateEmail)
			},
			expectedErrCode: apperrors.Conflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := beginTestTransaction(t, db)
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewUserRepository(tx)
			err := repo.CreateUser(context.Background(), tt.user)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)

				// データベースに実際に保存されているか確認
				var count int
				query := "SELECT COUNT(*) FROM users WHERE email = $1"
				err = tx.Get(&count, query, tt.user.Email)
				if err != nil {
					t.Errorf("ユーザー存在確認に失敗しました: %v", err)
				}
				if count != 1 {
					t.Errorf("expected user count 1, got %d", count)
				}

				if tt.assertion != nil {
					tt.assertion(t, tt.user)
				}
			}
		})
	}
}

// TestGetUserByEmail - メールアドレスによるユーザー取得のテスト
func TestGetUserByEmail(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		email           string
		setup           func(*sqlx.Tx) *models.User
		expectedErrCode apperrors.ErrCode
		assertion       func(*testing.T, *models.User, *models.User)
	}{
		{
			name:  "正常系: 存在するメールアドレスでユーザーを取得できる",
			email: testEmail2,
			setup: func(tx *sqlx.Tx) *models.User {
				return createTestUserWithEmail(t, tx, testEmail2)
			},
			assertion: func(t *testing.T, expected, got *models.User) {
				if got.UserID != expected.UserID {
					t.Errorf("UserID: expected %d, got %d", expected.UserID, got.UserID)
				}
				if got.Email != expected.Email {
					t.Errorf("Email: expected %s, got %s", expected.Email, got.Email)
				}
				if got.Role != expected.Role {
					t.Errorf("Role: expected %s, got %s", expected.Role, got.Role)
				}
				if got.CreatedAt.IsZero() {
					t.Error("CreatedAt が設定されていません")
				}
				if got.UpdatedAt.IsZero() {
					t.Error("UpdatedAt が設定されていません")
				}
			},
		},
		{
			name:            "異常系: 存在しないメールアドレスの場合はNoDataエラーを返す",
			email:           nonExistentEmail,
			setup:           func(tx *sqlx.Tx) *models.User { return nil },
			expectedErrCode: apperrors.NoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := beginTestTransaction(t, db)
			defer tx.Rollback()

			expectedUser := tt.setup(tx)

			repo := repositories.NewUserRepository(tx)
			got, err := repo.GetUserByEmail(context.Background(), tt.email)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
				// エラーケースでは空のユーザーが返される
				if got.UserID != 0 || got.Email != "" {
					t.Errorf("エラーケースで空でないユーザーが返されました: %+v", got)
				}
			} else {
				assertNoError(t, err)

				if expectedUser == nil {
					t.Fatal("expectedUser is nil")
				}

				if tt.assertion != nil {
					tt.assertion(t, expectedUser, &got)
				}
			}
		})
	}
}
