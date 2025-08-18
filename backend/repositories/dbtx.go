package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DBTX interface {
	// 1行取得 (SELECT ... LIMIT 1)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 複数行取得 (SELECT)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	// 複数行取得 (手動Scan用)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) // 👈 この行を追加
	// 1行取得 (Scan用)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	// 登録・更新・削除 (INSERT, UPDATE, DELETE)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// プリペアドステートメント
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	// IN句のプレースホルダ置換
	Rebind(query string) string
}
