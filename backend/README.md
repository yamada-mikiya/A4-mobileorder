# Mobile Order Backend

モバイルオーダーシステムのバックエンドAPIサーバーです。Go言語とPostgreSQLを使用して構築されています。

## 🚀 機能概要

- **ユーザー管理**: 認証・認可機能
- **店舗管理**: 店舗情報の管理
- **商品管理**: 商品カタログの管理
- **注文処理**: リアルタイムな注文管理
- **管理者機能**: 店舗スタッフ向けの管理画面

## 🛠 技術スタック

- **言語**: Go
- **フレームワーク**: Echo
- **データベース**: PostgreSQL
- **ORM**: sqlx
- **API文書**: Swagger
- **コンテナ**: Docker & Docker Compose
- **マイグレーション**: golang-migrate

## ⚡ クイックスタート

### 1. リポジトリのクローン

```bash
git clone https://github.com/A4-dev-team/mobileorder.git
cd mobileorder/backend
```

### 2. 環境設定

```bash
# 環境変数ファイルをコピー
cp .env.example .env

# 必要に応じて .env ファイルを編集
# デフォルト設定でも動作します
```

### 3. 開発環境の起動

```bash
# 初回起動（ビルド込み）
docker compose up --build

# 2回目以降の起動
docker compose up
```

### 4. APIの確認

- **API サーバー**: http://localhost:8080
- **Swagger UI**: http://localhost
- **ヘルスチェック**: http://localhost:8080/health

### 5. API使用例

#### 認証関連

```bash
# ケース1: 新規ユーザーサインアップ（メールアドレスのみ）
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "new.user@example.com"}'

# ケース2: ゲスト注文後のサインアップ（注文引き継ぎ）
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "guest.shopper@example.com",
    "guest_order_token": "15ff4999-2cfd-41f3-b744-926e7c5c7a0e"
  }'

# ケース1: 既存ユーザーログイン
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin1@example.com"}'

# ケース2: ゲスト注文後のログイン（注文引き継ぎ）
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "existing.user@example.com",
    "guest_order_token": "15ff4999-2cfd-41f3-b744-926e7c5c7a0e"
  }'
```

#### 商品・店舗関連

```bash
# 商品一覧取得（店舗ID: 1）
curl http://localhost:8080/shops/1/products

# 店舗情報取得
curl http://localhost:8080/shops/1
```

#### 注文関連

```bash
# 認証トークンが無効な場合（ゲスト注文）
curl -X POST http://localhost:8080/shops/1/guest-orders \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"item_id": 1, "quantity": 2},
      {"item_id": 2, "quantity": 1}
    ]
  }'

# 認証トークンが有効な場合（ユーザー注文）
curl -X POST http://localhost:8080/shops/1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "items": [
      {"item_id": 1, "quantity": 2},
      {"item_id": 2, "quantity": 1}
    ]
  }'

# 注文履歴取得（認証必要）
curl http://localhost:8080/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 特定注文のステータス確認（認証必要）
curl http://localhost:8080/orders/6/status \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 注文削除（認証必要）
curl -X DELETE http://localhost:8080/orders/6/delete \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### 管理者機能

```bash
# 調理中注文一覧取得（管理者権限必要）
curl http://localhost:8080/admin/shops/1/orders/cooking \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"

# 完了済み注文一覧取得（管理者権限必要）
curl http://localhost:8080/admin/shops/1/orders/completed \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"

# 注文ステータス更新（cooking → completed → handed）
curl -X PUT http://localhost:8080/admin/orders/6/status \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"

# 商品在庫状態更新
curl -X PUT http://localhost:8080/admin/products/1/availability \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -d '{"available": false}'
```

## API エンドポイント一覧

### 認証
- `POST /auth/signup` - ユーザー登録
- `POST /auth/login` - ユーザーログイン

### 店舗・商品
- `GET /shops/:shop_id` - 店舗情報取得
- `GET /shops/:shop_id/products` - 商品一覧取得

### 注文（認証不要）
- `POST /shops/:shop_id/guest-orders` - ゲスト注文作成

### 注文（認証必要）
- `POST /shops/:shop_id/orders` - ユーザー注文作成
- `GET /orders` - 注文履歴取得
- `GET /orders/:order_id/status` - 注文ステータス確認
- `DELETE /orders/:order_id/delete` - 注文削除

### 管理者機能（管理者権限必要）
- `GET /admin/shops/:shop_id/orders/cooking` - 調理中注文一覧
- `GET /admin/shops/:shop_id/orders/completed` - 完了済み注文一覧
- `PUT /admin/orders/:order_id/status` - 注文ステータス更新
- `PUT /admin/products/:product_id/availability` - 商品在庫更新

## 開発ガイド

### 認証システム

このAPIは**メールアドレスのみでの認証**と**ゲスト注文の引き継ぎ機能**を提供します。

#### 認証フロー

1. **新規ユーザー**: メールアドレスでサインアップ → JWTトークン取得
2. **既存ユーザー**: メールアドレスでログイン → JWTトークン取得
3. **ゲストユーザー**: 注文作成 → `guest_order_token` 取得 → サインアップ/ログイン時に注文引き継ぎ

#### 認証が必要なエンドポイント

以下のエンドポイントでは `Authorization: Bearer <JWT_TOKEN>` ヘッダーが必要です：

- `GET /orders` - 注文履歴取得
- `GET /orders/:order_id/status` - 注文ステータス確認
- `DELETE /orders/:order_id/delete` - 注文削除
- `GET /admin/shops/:shop_id/orders/cooking` - 調理中注文一覧（管理者）
- `GET /admin/shops/:shop_id/orders/completed` - 完了済み注文一覧（管理者）
- `PUT /admin/orders/:order_id/status` - 注文ステータス更新（管理者）
- `PUT /admin/products/:product_id/availability` - 商品在庫更新（管理者）

### 環境変数

#### アプリケーション設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|------------|
| `DATABASE_URL` | PostgreSQL接続URL | `postgres://myuser:mypassword@db:5432/mydb?sslmode=disable` |
| `PORT` | APIサーバーポート | `8080` |
| `SECRET` | JWT秘密鍵 | `mobileorder` |

#### データベースコンテナ設定

| 変数名 | 説明 | デフォルト値 |
|--------|------|------------|
| `POSTGRES_USER` | PostgreSQLユーザー名 | `myuser` |
| `POSTGRES_PASSWORD` | PostgreSQLパスワード | `mypassword` |
| `POSTGRES_DB` | データベース名 | `mydb` |

#### .env ファイル例

```bash
# アプリからDBに接続する
DATABASE_URL=postgres://myuser:mypassword@db:5432/mydb?sslmode=disable
PORT=8080
SECRET=mobileorder

# DBコンテナの初期化
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=mydb
```

> ⚠️ **セキュリティ注意**: 本番環境では強力なパスワードとランダムなSECRETキーを使用してください。

### API文書の更新

```bash
# Swagger文書を再生成
swag init

# サーバーを再起動
docker compose up --build
```

### データベース操作

```bash
# データベースに直接接続
docker compose exec db psql -U $POSTGRES_USER -d $POSTGRES_DB
# または具体的に
docker compose exec db psql -U myuser -d mydb

# マイグレーション状態確認
docker compose exec app migrate -path ./db/migrations -database "$DATABASE_URL" version

# マイグレーション実行
docker compose exec app migrate -path ./db/migrations -database "$DATABASE_URL" up

# マイグレーションロールバック
docker compose exec app migrate -path ./db/migrations -database "$DATABASE_URL" down 1
```

## テスト

本プロジェクトでは、Makefileを使用して効率的にテストを実行できます。すべてのテストでカバレッジレポートが自動生成されます。

### 🚀 クイック実行

```bash
# 全層のテストを順次実行（推奨）
make test

# 利用可能なコマンド一覧を表示
make help
```

### 📋 層別テスト

```bash
# コントローラー層のテスト
make test-controllers

# サービス層のテスト（ユニット + 結合テスト）
make test-services

# リポジトリ層のテスト
make test-repositories

# 結合テストのみ実行
make test-integration
```

### 📊 カバレッジレポート

```bash
# カバレッジレポート生成（HTMLファイル）
make test-coverage

# カバレッジサマリー表示
make coverage-summary

# テスト結果ファイルのクリーンアップ
make clean
```

### 🔧 Docker環境での実行

Docker環境でテストを実行する場合：

```bash
# Docker コンテナ内でテスト実行
docker compose exec app make test

# 特定の層のみテスト
docker compose exec app make test-repositories

# カバレッジレポート生成
docker compose exec app make test-coverage
```

### 📈 カバレッジレポート確認

テスト実行後、`coverage/`ディレクトリに以下のファイルが生成されます：

- `coverage/report.html` - HTMLフォーマットの詳細レポート（ブラウザで開く）
- `coverage/combined.out` - 統合カバレッジデータ
- `coverage/controllers.out` - コントローラー層の個別カバレッジ
- `coverage/services.out` - サービス層の個別カバレッジ  
- `coverage/repositories.out` - リポジトリ層の個別カバレッジ

```bash
# HTMLレポートをブラウザで開く（例：Linux）
xdg-open coverage/report.html

# HTMLレポートをブラウザで開く（例：macOS）
open coverage/report.html
```

### 🛠 従来のGoコマンド（参考）

Makefileを使わない場合の従来のコマンド：

```bash
# 全モジュールのテスト実行
go test -v ./...

# リポジトリ層のテスト（カバレッジ付き）
DATABASE_URL="postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable" \
go test -v -cover -coverpkg=./repositories -coverprofile=repositories_coverage.out ./repositories

# カバレッジレポート生成
go tool cover -html=repositories_coverage.out -o repositories_coverage.html
```

## 🗂 プロジェクト構造

```
backend/
├── api/                    # APIルーティング
│   ├── router.go
│   └── middlewares/        # ミドルウェア
├── controllers/            # HTTPハンドラー
├── services/              # ビジネスロジック
├── repositories/          # データアクセス層
├── models/                # データモデル
├── db/                    # データベース関連
│   └── migrations/        # マイグレーションファイル
├── docs/                  # Swagger生成ファイル
├── apperrors/             # エラーハンドリング
├── validators/            # バリデーション
├── connectDB/             # DB接続設定
├── docker-compose.yml     # Docker設定
├── Dockerfile            # Dockerイメージ定義
└── main.go               # エントリーポイント
```

## 🐳 Docker コマンド

```bash
# 開発環境起動
docker compose up --build

# バックグラウンド起動
docker compose up -d

# コンテナ停止
docker compose down

# コンテナ停止 + データベースリセット
docker compose down -v
```