# 開発環境の構築
## バックエンド
### 0. envファイルの作成
.emvファイルを作る
.env.exampleをコピーしたやつを.envファイルに移す。
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
docker compose exec db psql -U myuser -d mydb
#userとdbnameは自分の.envファイルの設定を見て書き換えてください。.envファイルの例は.env.exampleにあります。
```
### extra.1 swagger
#swaggerの変更時
```bash
swag init
docker compose up --build
#変更がなくて初めてswaggerを生成する場合はdocker compose up --buildで大丈夫です。
#swaggerのURLはhttp://localhostをブラウザにぶち込む

