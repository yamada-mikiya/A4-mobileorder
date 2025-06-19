# 開発環境の構築
## バックエンド
### 1. データベースコンテナの起動
```bash
cd backend
docker compose up --build
#2回目以降の起動やファイルの変更がないときは次でも大丈夫です
cd backend
docker compose up
```
### 2.Docker環境を落とす 
```bash
# コンテナを終了（データベースの中身は残す）
docker compose down

# コンテナ + データベースごと削除
docker compose down -v
```

### 3.データベースに接続してテーブル情報を見たいとき
```bash
docker compose exec db psql -U user -d dbname
#userとdbnameは自分の.envファイルの設定を見て書き換えてください。.envファイルの例は.env.exampleにあります。
```
### http://localhost:8080/auth/login　と　http://localhost:8080/auth/login　をポストマンで実行するには
```bash
# POSTメソッドで行う。JSON形式でリクエストボディに下のように書く。(Content-Typeはapplication/jsonであることに注意)
{
    "email": "test-user@example.com"
}
```
## フロントエンド