package services_test

import (
	"errors"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
)

// assertAppError - エラーアサーションヘルパー関数（リポジトリ層と統一）
// 指定されたエラーコードのAppErrorが発生したかを検証します
func assertAppError(t *testing.T, err error, expectedCode apperrors.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("期待したエラーコード '%s' が返されませんでしたが、エラーが発生しませんでした", expectedCode)
	}
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("期待したエラータイプは *apperrors.AppError ですが、実際は %T でした", err)
	}
	if appErr.ErrCode != expectedCode {
		t.Fatalf("期待したエラーコード '%s' ですが、実際は '%s' でした", expectedCode, appErr.ErrCode)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}
}
