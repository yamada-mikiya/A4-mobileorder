package validators

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

type Validator[T any] interface {
	Validate(t T) error
}

type customValidator[T any] struct {
	validate *validator.Validate
}

var (
	singletonValidator *validator.Validate
	once               sync.Once
)

func getSingletonValidator() *validator.Validate {
	once.Do(func() {
		singletonValidator = validator.New()
	})
	return singletonValidator
}

func NewValidator[T any]() Validator[T] {
	return &customValidator[T]{validate: getSingletonValidator()}
}

func (v *customValidator[T]) Validate(t T) error {
	return v.validate.Struct(t)
}
