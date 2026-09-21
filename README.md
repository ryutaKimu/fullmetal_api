# 本リポジトリについて

[![CI](https://github.com/v420v/fullmetal_api/actions/workflows/ci.yml/badge.svg)](https://github.com/v420v/fullmetal_api/actions/workflows/ci.yml)

こちらは鋼の錬金術師のナレーションを返す非公式APIです。
あくまでファンメイドであるため、公式とは一切関係ありません。

開発を始めた理由や仕様は `docs/` 配下にまとまっています。

| ファイル | 内容 |
| --- | --- |
| `docs/plan.md` | なぜ作るのか、利用シーン、拡張案 |
| `docs/requirement.md` | エンドポイント・レスポンス・データ項目の仕様 |
| `docs/archtecture.md` | 全体の設計方針とディレクトリ構成 |
| `docs/faq.md` | 仕様やデータについての設計判断の理由 |

# 技術構成

Go 1.26.4（標準ライブラリのみ。外部依存なし）
データは `data/narrations.json` を `go:embed` でバイナリに埋め込んでいます。

# 使い方

https://fullmetalapi.vercel.app/v1/narrations/random にアクセス

## エンドポイント

APIはパスの先頭でバージョンを指定します（現行バージョンは `v1`）。

| メソッド | パス | 説明 |
| --- | --- | --- |
| GET | `/v1/narrations/random` | ナレーションを1件ランダムに返す |
| GET | `/v1/narrations/{episode}` | エピソードで指定されたナレーションを返す |

現在はこの2つのみです。ルート（`/`）を含む他のパスは 404 を返します。

バージョン導入前の旧パス `/narrations/...` へのアクセスは、互換性のため対応する `/v1/narrations/...` へ 308 リダイレクトされます（クエリパラメータも引き継がれます）。新規の利用では `/v1` を直接指定してください。

## パスパラメータ

| 名前 | 値 | 説明 |
| --- | --- | --- |
| `episode` | `1`〜`63` | 取得するエピソード。整数でない場合は 400、該当データがない場合は 404 |

## クエリパラメータ

| 名前 | 必須 | 値 | 既定値 | 説明 |
| --- | --- | --- | --- | --- |
| `lang` | 任意 | `ja` / `en` | `ja` | 返却する言語。対応外の値は 400 |

## レスポンス

データがある場合（200）

```json
{
  "data": {
    "episode": 1,
    "title": "鋼の錬金術師",
    "narrations": [
      "リゼンブール、そこは少年達が生まれ、母と供に過ごした優しい町",
      "失った笑顔を求め、少年達は禁忌を犯し、真理を目撃する",
      "次回、鋼の錬金術師 FULLMETAL ALCHEMIST 第2話『はじまりの日』",
      "旅立ちを決めたのは、自分の心"
    ]
  }
}
```

`/v1/narrations/random` で該当データがない場合（200）

```json
{ "data": null }
```

`/v1/narrations/{episode}` で該当データがない場合（404）

```json
{ "data": null, "error": "narration not found" }
```

エピソードが存在しない場合と、存在しても指定言語が未翻訳の場合は、どちらもこの 404 になります。

`episode` が整数でない場合（400）

```json
{ "data": null, "error": "episode must be an integer" }
```

対応していない言語を指定した場合（400）

```json
{ "data": null, "error": "the language is not supported" }
```

`/v1/narrations/random` がエラーではなく 200 + `data: null` を返す理由は `docs/faq.md` に記載しています。

# データの収録状況

- 全63件（episode 1〜63）
- **日本語（`ja`）のみ収録済み。英語（`en`）は全件未翻訳（null）です。**

`lang` 指定時は title と narrations の両方が非nullのデータだけを対象にするため、現状 `?lang=en` は `/v1/narrations/random` なら 200 + `{ "data": null }`、`/v1/narrations/{episode}` なら 404 を返します。英訳は今後、段階的に追加していく予定です。

# ローカルでの実行

```sh
go run ./cmd/api
# => listening on :9090

curl "http://localhost:9090/v1/narrations/random"
curl "http://localhost:9090/v1/narrations/random?lang=en"
curl "http://localhost:9090/v1/narrations/5"
```

ポートは環境変数 `PORT` で変更できます（未指定時は `9090`）。

テストの実行:

```sh
go test ./...
```

`main` への push と Pull Request では GitHub Actions（`.github/workflows/ci.yml`）が gofmt・go vet・go test（`-race`）を自動実行します。

# データを追加・修正するには

1. `data/narrations.json` を編集する
2. 値の制約（episodeは1〜63の整数で重複不可、未翻訳は `null`、空文字・空配列は不可）は `docs/requirement.md` に従う
3. `go test ./...` でデータの検証テストを通す
4. JSONは `go:embed` で埋め込まれているため、反映には再ビルドが必要です

# 権利について

本APIはファンによる非公式のプロジェクトであり、原作および映像作品の権利者とは一切関係ありません。

収録しているナレーションは、TVアニメ『鋼の錬金術師 FULLMETAL ALCHEMIST』の台詞を引用したものです。原作・映像作品に関する権利はすべて権利者に帰属します。

本リポジトリは非営利の個人利用・学習目的で公開しています。

なお、ソースコード部分のライセンスは未設定です（LICENSE ファイルは配置していません）。
