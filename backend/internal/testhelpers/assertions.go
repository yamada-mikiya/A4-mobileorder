package testhelpers

import (
	"errors"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
)

// AssertAppError は、エラーが期待されるAppErrorかどうかを検証するヘルパー関数です。
// 全てのテストファイルで統一的に使用できます。
func AssertAppError(t *testing.T, err error, expectedCode apperrors.ErrCode) {
	t.Helper()
	if err == nil {
		t.Errorf("エラーが期待されますが、nilが返されました")
		return
	}

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.ErrCode != expectedCode {
			t.Errorf("期待されるエラーコード = %s, 実際 = %s", expectedCode, appErr.ErrCode)
		}
	} else {
		t.Errorf("期待されるAppError型ではありません: %v", err)
	}
}

// AssertNoError は、エラーが発生しないことを検証するヘルパー関数です。
// 全てのテストファイルで統一的に使用できます。
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("エラーは期待されませんが、エラーが発生しました: %v", err)
	}
}
