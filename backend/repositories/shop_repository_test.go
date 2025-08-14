package repositories_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

const (
	testAdminUserID1   = 101
	testStaffShopID1   = 101
	testStaffShopID2   = 102
	nonExistentAdminID = 999
	multiShopAdminID   = 104
)

// createTestShopStaff - テスト用の店舗スタッフ関係を作成するヘルパー関数
func createTestShopStaff(t *testing.T, tx *sqlx.Tx, userID int, shopID int) {
	t.Helper()

	// ユーザーを作成（重複エラーを無視）
	userQuery := `INSERT INTO users (user_id, email) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`
	_, err := tx.Exec(userQuery, userID, fmt.Sprintf("user%d@test.com", userID))
	if err != nil {
		t.Fatalf("テスト用ユーザーの作成に失敗しました: %v", err)
	}

	// 店舗を作成（重複エラーを無視）
	shopQuery := `INSERT INTO shops (shop_id, name) VALUES ($1, $2) ON CONFLICT (shop_id) DO NOTHING`
	_, err = tx.Exec(shopQuery, shopID, fmt.Sprintf("Test Shop %d", shopID))
	if err != nil {
		t.Fatalf("テスト用店舗の作成に失敗しました: %v", err)
	}

	// shop_staff テーブルに関係を挿入
	query := `INSERT INTO shop_staff (user_id, shop_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err = tx.Exec(query, userID, shopID)
	if err != nil {
		t.Fatalf("テスト用店舗スタッフ関係の作成に失敗しました: %v", err)
	}
}

// TestFindShopIDByAdminID - 管理者IDから店舗ID取得のテスト
func TestFindShopIDByAdminID(t *testing.T) {
	db := NewTestDB(t)

	tests := []struct {
		name            string
		userID          int
		setup           func(*sqlx.Tx)
		want            int
		expectedErrCode apperrors.ErrCode
	}{
		{
			name:   "正常系: 管理者に紐づく店舗が1つの場合は店舗IDを返す",
			userID: testAdminUserID1,
			setup: func(tx *sqlx.Tx) {
				createTestShopStaff(t, tx, testAdminUserID1, testStaffShopID1)
			},
			want: testStaffShopID1,
		},
		{
			name:            "異常系: 管理者に紐づく店舗が存在しない場合はNoDataエラーを返す",
			userID:          nonExistentAdminID,
			setup:           func(tx *sqlx.Tx) {},
			want:            0,
			expectedErrCode: apperrors.NoData,
		},
		{
			name:   "異常系: 管理者に複数店舗が紐づく場合はUnknownエラーを返す",
			userID: multiShopAdminID,
			setup: func(tx *sqlx.Tx) {
				createTestShopStaff(t, tx, multiShopAdminID, testStaffShopID1)
				createTestShopStaff(t, tx, multiShopAdminID, testStaffShopID2)
			},
			want:            0,
			expectedErrCode: apperrors.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.MustBegin()
			defer tx.Rollback()

			tt.setup(tx)

			repo := repositories.NewShopRepository(tx)
			got, err := repo.FindShopIDByAdminID(context.Background(), tt.userID)

			if tt.expectedErrCode != "" {
				assertAppError(t, err, tt.expectedErrCode)
			} else {
				assertNoError(t, err)
				if got != tt.want {
					t.Errorf("FindShopIDByAdminID() = %d, want %d", got, tt.want)
				}
			}
		})
	}
}
