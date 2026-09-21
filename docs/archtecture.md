# 目的
全体の設計方針についてはこちらに記載。

# ディレクトリ構成
```
.
├── cmd/
│   └── api/
│       └── main.go          # 起動処理
├── internal/
│   ├── router/              # パスとバージョン（/v1）の対応づけ
│   ├── handler/             # リクエスト受付・レスポンス
│   ├── service/             # 絞り込み・ランダム選択
│   ├── repository/          # JSON読み込み
│   └── model/               # データの型定義
└── data/
    └── narrations.json
```

