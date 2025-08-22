package services

import (
	"context"

	"github.com/A4-dev-team/mobileorder.git/apperrors"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/jmoiron/sqlx"
)

// TransactionManager データベーストランザクションを管理するインターフェース
type TransactionManager interface {
	WithOrderTransaction(ctx context.Context, fn func(repositories.OrderRepository) error) error
	WithUserOrderTransaction(ctx context.Context, fn func(repositories.UserRepository, repositories.OrderRepository) error) error
}

// sqlxTransactionManager sqlxを使用してTransactionManagerを実装する構造体
type sqlxTransactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager 新しいトランザクションマネージャーを作成する
func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &sqlxTransactionManager{db: db}
}

// WithOrderTransaction 注文リポジトリを使用してトランザクション内で関数を実行する
func (tm *sqlxTransactionManager) WithOrderTransaction(ctx context.Context, fn func(repositories.OrderRepository) error) (err error) {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = apperrors.Unknown.Wrap(rbErr, "パニック発生時のトランザクションロールバックに失敗しました。")
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = apperrors.Unknown.Wrapf(rbErr, "エラー発生時のトランザクションロールバックに失敗しました。元のエラー: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				err = apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
			}
		}
	}()

	txRepo := repositories.NewOrderRepository(tx)
	return fn(txRepo)
}

// WithUserOrderTransaction ユーザーと注文リポジトリを使用してトランザクション内で関数を実行する
func (tm *sqlxTransactionManager) WithUserOrderTransaction(ctx context.Context, fn func(repositories.UserRepository, repositories.OrderRepository) error) (err error) {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return apperrors.Unknown.Wrap(err, "トランザクションの開始に失敗しました。")
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = apperrors.Unknown.Wrap(rbErr, "パニック発生時のトランザクションロールバックに失敗しました。")
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = apperrors.Unknown.Wrapf(rbErr, "エラー発生時のトランザクションロールバックに失敗しました。元のエラー: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				err = apperrors.Unknown.Wrap(err, "トランザクションのコミットに失敗しました。")
			}
		}
	}()

	txUserRepo := repositories.NewUserRepository(tx)
	txOrderRepo := repositories.NewOrderRepository(tx)
	return fn(txUserRepo, txOrderRepo)
}
