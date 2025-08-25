# =================================
# Production Dockerfile
# =================================

# ビルドステージ: アプリケーションをビルド
FROM golang:1.24-alpine AS builder

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates tzdata curl tar

# migrate バイナリをダウンロード
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# ワーキングディレクトリを設定
WORKDIR /app

# Go modules をコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# CGO無効でバイナリをビルド（静的リンク）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main .

# =================================
# 本番実行ステージ: 最小限のイメージ
FROM alpine:latest

# 必要最小限のパッケージをインストール
RUN apk add --no-cache ca-certificates tzdata

# 非rootユーザーを作成
RUN adduser -D -s /bin/sh appuser

# アプリケーションバイナリとマイグレーションファイルをコピー
COPY --from=builder /app/main /app/main
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/db/migrations /app/db/migrations

# 非rootユーザーに切り替え
USER appuser

# ワーキングディレクトリを設定
WORKDIR /app

# ポートを公開
EXPOSE 8080

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/main", "-health-check"] || exit 1

# アプリケーションを実行
ENTRYPOINT ["/app/main"]
