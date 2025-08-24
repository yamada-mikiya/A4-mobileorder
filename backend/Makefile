# mobileorder backend Makefile

.PHONY: test test-controllers test-services test-repositories test-validators test-coverage test-integration clean help setup-mockdata

# デフォルトターゲット
all: test

# mockDataの再挿入（Docker DB用）
setup-mockdata:
	@echo "🗄️  Docker DBにmockDataを挿入..."
	@if ! docker compose ps | grep -q "postgres-db.*Up"; then \
		echo "❌ Docker Compose DBが起動していません。'docker compose up -d' を実行してください。"; \
		exit 1; \
	fi
	@cat mockData.sql | docker compose exec -T db psql -U myuser -d mydb
	@echo "✅ mockData挿入完了"

# 全てのテストを順次実行
test:
	@echo "=== mobileorder プロジェクト全体テスト実行 ==="
	@$(MAKE) clean
	@$(MAKE) test-validators
	@$(MAKE) test-controllers
	@$(MAKE) test-services
	@$(MAKE) test-repositories
	@$(MAKE) test-coverage
	@echo "=== テスト実行完了 ==="

# バリデーター層のテスト
test-validators:
	@echo "🧪 バリデーター層のテスト実行..."
	@mkdir -p coverage
	go test ./validators -v -coverprofile=coverage/validators.out -covermode=atomic
	@echo "✅ バリデーター層: 完了"

# コントローラー層のテスト
test-controllers:
	@echo "🧪 コントローラー層のテスト実行..."
	@mkdir -p coverage
	go test ./controllers -v -coverprofile=coverage/controllers.out -covermode=atomic
	@echo "✅ コントローラー層: 完了"

# サービス層の全テスト（ユニット+結合）
test-services:
	@echo "🧪 サービス層の全テスト実行..."
	@mkdir -p coverage
	@echo "📥 データベースにmockDataを挿入中..."
	@$(MAKE) setup-mockdata > /dev/null 2>&1 || echo "⚠️  mockData挿入をスキップ（Docker未起動またはDBエラー）"
	@export DATABASE_URL="postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" && \
	go test ./services -v -coverprofile=coverage/services.out -covermode=atomic
	@echo "✅ サービス層: 完了"

# リポジトリ層のテスト
test-repositories:
	@echo "🧪 リポジトリ層のテスト実行..."
	@mkdir -p coverage
	@echo "📥 データベースにmockDataを挿入中..."
	@$(MAKE) setup-mockdata > /dev/null 2>&1 || echo "⚠️  mockData挿入をスキップ（Docker未起動またはDBエラー）"
	@export DATABASE_URL="postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" && \
	go test ./repositories -v -coverprofile=coverage/repositories.out -covermode=atomic
	@echo "✅ リポジトリ層: 完了"

# カバレッジレポート生成
test-coverage:
	@echo "📊 カバレッジレポート生成..."
	@mkdir -p coverage
	@echo "mode: atomic" > coverage/combined.out
	@if [ -f coverage/validators.out ]; then tail -n +2 coverage/validators.out >> coverage/combined.out; fi
	@if [ -f coverage/controllers.out ]; then tail -n +2 coverage/controllers.out >> coverage/combined.out; fi
	@if [ -f coverage/services.out ]; then tail -n +2 coverage/services.out >> coverage/combined.out; fi
	@if [ -f coverage/repositories.out ]; then tail -n +2 coverage/repositories.out >> coverage/combined.out; fi
	go tool cover -html=coverage/combined.out -o coverage/report.html
	@echo "📁 HTMLレポート: coverage/report.html"
	@echo "📊 全体のカバレッジ:"
	@go tool cover -func=coverage/combined.out | tail -1

# カバレッジ結果のみ表示
coverage-summary:
	@if [ -f coverage/combined.out ]; then \
		echo "📊 全体のカバレッジサマリー:"; \
		go tool cover -func=coverage/combined.out | tail -1; \
	else \
		echo "❌ カバレッジファイルが見つかりません。先に 'make test' を実行してください。"; \
	fi

# 統合テストのみ実行
test-integration:
	@echo "🧪 統合テスト実行..."
	@echo "📥 データベースにmockDataを挿入中..."
	@$(MAKE) setup-mockdata > /dev/null 2>&1 || echo "⚠️  mockData挿入をスキップ（Docker未起動またはDBエラー）"
	@export DATABASE_URL="postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" && \
	echo "🔗 データベース接続URL: $$DATABASE_URL" && \
	go test -run="Integration" ./services -v
	@echo "✅ 統合テスト: 完了"

# テスト結果ファイルのクリーンアップ
clean:
	@echo "🧹 テスト結果ファイルをクリーンアップ..."
	rm -rf coverage/
	rm -f c.out coverage.html *_coverage.out

# ヘルプ
help:
	@echo "使用可能なコマンド:"
	@echo "  make test              - 全てのテストを順次実行（推奨）"
	@echo "  make setup-mockdata    - Docker DBにmockDataを再挿入"
	@echo "  make test-validators   - バリデーター層のみテスト"
	@echo "  make test-controllers  - コントローラー層のみテスト"
	@echo "  make test-services     - サービス層のみテスト"
	@echo "  make test-repositories - リポジトリ層のみテスト"
	@echo "  make test-integration  - 統合テストのみ実行"
	@echo "  make test-coverage     - カバレッジレポート生成"
	@echo "  make coverage-summary  - カバレッジサマリー表示"
	@echo "  make clean             - テスト結果ファイルをクリーンアップ"
	@echo "  make help              - このヘルプを表示"
