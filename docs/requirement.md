# 機能要件

## エンドポイント

GET /v1/narrations/random
lang（任意）：ja / en、省略時は ja
対応外の値は400エラー

## レスポンス値

+ データがある場合
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

+ データがない場合
```json
{ "data": null }
```

+ 400エラーの場合(言語が対応していない時)
```json
{ "data": null, "error": "the language is not supported" }
```

## エンドポイント（複数件のランダム取得）

GET /v1/narrations/random/{episodes}
episodes（必須）：取得する件数。1 ~ 63の整数
lang（任意）：ja / en、省略時は ja
対応外の値は400エラー

- 指定言語でtitleとnarrationsの両方が非nullのデータから、episodes件を重複なしでランダムに返す
- 並び順はランダムとする
- 指定言語で取得できる件数がepisodesより少ない場合は、足りる分だけを返さず404エラーとする

## レスポンス値（複数件のランダム取得）

+ データがある場合（dataは配列。件数はepisodesと一致する）
```json
{
  "data": [
    {
      "episode": 1,
      "title": "鋼の錬金術師",
      "narrations": [
        "リゼンブール、そこは少年達が生まれ、母と供に過ごした優しい町",
        "失った笑顔を求め、少年達は禁忌を犯し、真理を目撃する",
        "次回、鋼の錬金術師 FULLMETAL ALCHEMIST 第2話『はじまりの日』",
        "旅立ちを決めたのは、自分の心"
      ]
    }
  ]
}
```

+ 404エラーの場合(指定言語で取得できる件数がepisodesより少ない時)
```json
{ "data": null, "error": "narration not found" }
```

+ 400エラーの場合(episodesが整数でない時)
```json
{ "data": null, "error": "episode must be an integer" }
```

+ 400エラーの場合(episodesが1 ~ 63の範囲外の時)
```json
{ "data": null, "error": "episodes must be between 1 and 63" }
```

+ 400エラーの場合(言語が対応していない時)
```json
{ "data": null, "error": "the language is not supported" }
```

## エンドポイント（エピソード指定）

GET /v1/narrations/{episode}
episode（必須）：取得するエピソード。1 ~ 63の整数
lang（任意）：ja / en、省略時は ja
対応外の値は400エラー

- 指定したepisodeのナレーションを1件返す
- エピソードが存在しない場合と、存在しても指定言語が未翻訳（titleまたはnarrationsがnull）の場合は、区別せず404エラーとする
- 1 ~ 63の範囲外（0以下や64以上）も、存在しないエピソードとして404エラーとする（複数件のランダム取得のepisodesとは異なり、400にはしない）

## レスポンス値（エピソード指定）

+ データがある場合
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

+ 404エラーの場合(エピソードが存在しない、または指定言語が未翻訳の時)
```json
{ "data": null, "error": "narration not found" }
```

+ 400エラーの場合(episodeが整数でない時)
```json
{ "data": null, "error": "episode must be an integer" }
```

+ 400エラーの場合(言語が対応していない時)
```json
{ "data": null, "error": "the language is not supported" }
```

# データ項目

| 項目      | 型                 | 必須 | 説明                           |
| --------- | ------------------ | ---- | ------------------------------ |
| episode   | 数値               | ⚪︎   | ナレーションが使われた話数     |
| title     | 言語別の文字列     | ⚪︎   | 話タイトル|
| narrations | 言語別の文字列配列 | ⚪︎   | ナレーション本文               |

## 項目の意味とルール

- episodeはナレーションが使われた回を基準にすること
- titleはepisodeと同じ回のタイトルとする
- narrationsは読み上げの間ごとに区切り、配列の要素とする
- 日本語（ja）は必須とする
- 未翻訳の言語はnullとする
- 言語指定時は、titleとnarrationsの両方が非nullのデータのみを取得対象とする

# 値の制約

### episode
- 整数
- 1 ~ 63まで
- 重複不可（1話につき1件のため、一意の識別子を兼ねる）

### title
- ja：空でない文字列（必須）
- en：空でない文字列、またはnull

### narrations
- ja：1要素以上の文字列配列（必須）
- en：1要素以上の文字列配列、またはnull
- 各要素は空文字不可

# 保存方法
JSONファイルで管理