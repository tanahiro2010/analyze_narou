# Operations

## Local Run

`.env` を用意してから実行します。

```bash
go run ./cmd/bot daily
```

対応モード:

- `daily` / `d`
- `weekly` / `w`
- `quarterly` / `quarter` / `q`
- `yearly` / `annual` / `y`

引数なしの場合は `daily` です。

## Docker Compose

```bash
make run
```

または:

```bash
docker compose up --build
```

## GitHub Actions

### Test

`.github/workflows/test.yaml` は `push` / `pull_request` ごとに実行されます。

```bash
go test ./...
```

### Analyze

`.github/workflows/analyze.yaml` は手動実行と定期実行に対応しています。

手動実行では `daily` / `weekly` / `quarterly` / `yearly` を選択できます。

定期実行:

- 毎日 09:05 JST: daily
- 毎週月曜 09:15 JST: weekly
- 四半期初日 09:25 JST: quarterly
- 1月1日 09:35 JST: yearly

GitHub Actions で実行する場合は repository secrets を設定してください。

```text
DISCORD_WEBHOOK_URL
DENI_API_KEY
```

既存の `OPENAI_API_KEY` を使う場合は `DENI_API_KEY` の代わりに設定できます。

## Discord Output

Discord には embed 形式で投稿します。

実行開始時に、分析種別を H1 見出しとして先頭に投稿します。

```text
# デイリーランキング分析
```

ジャンル別分析では、以下のような項目が分かれて表示されます。

- 概要
- タイトル
- 紹介文
- ブックマーク
- 評価・文字数
- 上位タグ
- AI 要約
- AI タイトル・紹介分析
- AI タグ・ジャンル分析
- AI 読者反応分析
- AI 執筆アドバイス

全体分析も同じく embed 形式で投稿します。

## Performance

ジャンル別解析は並列実行されます。

```env
GENRE_ANALYZE_CONCURRENCY=4
```

値を大きくすると速くなりますが、なろう API、AI API、Discord Webhook への同時アクセスも増えます。レート制限や timeout が出る場合は値を下げてください。

## Troubleshooting

### Discord に投稿されない

- `DISCORD_WEBHOOK_URL` が正しいか確認
- Webhook の投稿先チャンネル権限を確認
- サーバー参加が必要な場合は承認制 Discord サーバーに参加

Discord 招待リンク: https://discord.gg/StYV8QMWPp

### AI 要約が出ない

- `DENI_API_KEY` または `OPENAI_API_KEY` が設定されているか確認
- `OPENAI_BASE_URL` が OpenAI API 互換の `/v1` base URL になっているか確認
- `OPENAI_MODEL` が利用可能なモデル ID か確認

### 実行が遅い

- `GENRE_ANALYZE_CONCURRENCY` を上げる
- AI API 側の応答速度を確認
- Discord Webhook の rate limit を確認

### なろう API の取得件数を変えたい

```env
NAROU_RANKING_LIMIT=100
```

値を大きくすると分析対象が増えますが、処理時間も増えます。
