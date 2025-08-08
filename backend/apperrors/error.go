package apperrors

import "fmt"

type AppError struct {
	ErrCode ErrCode `json:"err_code"`
	Message string  `json:"message"`
	Err     error   `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code ErrCode, message string, err error) error {
	return &AppError{
		ErrCode: code,
		Message: message,
		Err:     err,
	}
}

func (code ErrCode) Wrap(err error, message string) error {
	return New(code, message, err)
}

func (code ErrCode) Wrapf(err error, format string, a ...any) error {
	return New(code, fmt.Sprintf(format, a...), err)
}
