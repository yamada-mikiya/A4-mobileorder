package validators

import (
	"github.com/go-playground/validator/v10"
)

type Validator[T any] interface {
	Validate(t T) error
}

type customValidator[T any] struct {
	validate *validator.Validate
}

func NewValidator[T any]() Validator[T] {
	return &customValidator[T]{validate: validator.New()}
}

func (v *customValidator[T]) Validate(t T) error {
	return v.validate.Struct(t)
}
