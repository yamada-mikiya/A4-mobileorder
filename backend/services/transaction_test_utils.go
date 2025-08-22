package services

import (
	"context"

	"github.com/A4-dev-team/mobileorder.git/repositories"
)

// MockTransactionManager テスト用のシンプルなモック（トランザクションなしで関数を実行）
type MockTransactionManager struct {
	orderRepo repositories.OrderRepository
	userRepo  repositories.UserRepository
}

// NewMockTransactionManager テスト用の新しいモックトランザクションマネージャーを作成する
func NewMockTransactionManager(orderRepo repositories.OrderRepository) TransactionManager {
	return &MockTransactionManager{orderRepo: orderRepo}
}

// NewMockTransactionManagerFull テスト用の両方のリポジトリを持つ新しいモックトランザクションマネージャーを作成する
func NewMockTransactionManagerFull(userRepo repositories.UserRepository, orderRepo repositories.OrderRepository) TransactionManager {
	return &MockTransactionManager{
		userRepo:  userRepo,
		orderRepo: orderRepo,
	}
}

// WithOrderTransaction モックリポジトリで関数を実行する（実際のトランザクションなし）
func (m *MockTransactionManager) WithOrderTransaction(ctx context.Context, fn func(repositories.OrderRepository) error) error {
	return fn(m.orderRepo)
}

// WithUserOrderTransaction モックリポジトリで関数を実行する（実際のトランザクションなし）
func (m *MockTransactionManager) WithUserOrderTransaction(ctx context.Context, fn func(repositories.UserRepository, repositories.OrderRepository) error) error {
	return fn(m.userRepo, m.orderRepo)
}
