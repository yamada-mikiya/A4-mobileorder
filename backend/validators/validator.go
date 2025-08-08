package validators

import (
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(i interface{}) error
}

type customValidator struct {
	validator *validator.Validate
}

func NewValidator() Validator {
	return &customValidator{validator: validator.New()}
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
