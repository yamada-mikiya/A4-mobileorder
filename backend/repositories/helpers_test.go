package repositories_test

import (
	"errors"
	"testing"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
)

// assertAppError - エラーアサーションヘルパー関数
// 期待したエラーコードのAppErrorが実際の関数を実行することで発生したかを検証します
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

// assertNoError - エラーなしアサーションヘルパー関数
// 期待したエラーがないのに実際に関数を実行したときにエラーが発生していないことを検証します
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("予期しないエラーが発生しました: %v", err)
	}
}
