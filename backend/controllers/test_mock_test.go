package controllers_test

import (
	"github.com/stretchr/testify/mock"
)

// MockValidator は validators パッケージの Validator インターフェースのモック実装です
type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(i interface{}) error {
	args := m.Called(i)
	return args.Error(0)
}
