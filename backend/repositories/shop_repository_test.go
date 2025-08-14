package repositories_test

import (
	"context"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

// テスト定数
const (
	testAdminUserID1   = 101
	testStaffShopID1   = 1
	testStaffShopID2   = 2
	nonExistentAdminID = 999
	multiShopAdminID   = 104
)

// createTestShopStaff - テスト用の店舗スタッフ関係を作成するヘルパー関数
func createTestShopStaff(t *testing.T, tx *sqlx.Tx, userID int, shopID int) {
	t.Helper()

	// ユーザーが存在しない場合は作成
	createTestUserIfNotExists(t, tx, userID)

	// 店舗が存在しない場合は作成
	createTestShopIfNotExists(t, tx, shopID)

	// shop_staff テーブルに関係を挿入
	query := `INSERT INTO shop_staff (user_id, shop_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := tx.Exec(query, userID, shopID)
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
			tx := beginTestTransaction(t, db)
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
