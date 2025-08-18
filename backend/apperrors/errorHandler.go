package apperrors

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)
//最終的に返ってきたエラーをクライアントにJSON形式で見せてあげる
func ErrorHandler(err error, ctx echo.Context) {
	if ctx.Response().Committed {
		return
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		appErr = &AppError{
			ErrCode: Unknown,
			Message: "internal server error",
			Err:     err,
		}
	}

	slog.Error("error handled", "err_code", appErr.ErrCode, "detail", appErr.Error())

	var statusCode int
	switch appErr.ErrCode {
	case BadParam, ReqBodyDecodeFailed, ValidationFailed:
		statusCode = http.StatusBadRequest
	case Unauthorized:
		statusCode = http.StatusUnauthorized
	case Forbidden:
		statusCode = http.StatusForbidden
	case NoData:
		statusCode = http.StatusNotFound
	case Conflict:
		statusCode = http.StatusConflict
	default:
		statusCode = http.StatusInternalServerError
	}

	if err := ctx.JSON(statusCode, appErr); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}
