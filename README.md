# なろうアナライザー

小説家になろうのランキングを取得し、ジャンルごとの傾向を分析して Discord Webhook に送る bot です。

この bot の分析ログは、承認制の Discord サーバーに流す想定です。

Discord 招待リンク: https://discord.gg/StYV8QMWPp

## できること

- なろう小説 API (`novelapi`) から大ジャンル別ランキング相当の作品一覧を取得
- daily / weekly / quarterly / yearly のポイント順で分析
- タイトル、紹介文、タグ、評価、ブックマーク、文字数、連載状況などを集計
- OpenAI API 互換 API で AI 要約と執筆アドバイスを生成
- Discord Webhook に embed 形式で見やすく投稿
- ジャンル別解析を並列実行
- GitHub Actions で定期実行とテスト実行

## Requirements

- Go 1.26.4
- Docker / Docker Compose
- Discord Webhook URL
- OpenAI API 互換 API key

デフォルトでは Deni AI API を OpenAI API 互換ホストとして使います。

```env
OPENAI_BASE_URL=https://api.deniai.app/v1
OPENAI_MODEL=openai/gpt-5.2
```

## Quick Start

`.env` を作成します。

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
DENI_API_KEY=deni_xxx
```

ローカル実行:

```bash
go run ./cmd/bot daily
```

Docker Compose 実行:

```bash
make run
```

または:

```bash
docker compose up --build
```

## Ranking Mode

引数を省略すると `daily` で実行されます。

```bash
go run ./cmd/bot daily
go run ./cmd/bot weekly
go run ./cmd/bot quarterly
go run ./cmd/bot yearly
```

短縮形も使えます。

```bash
go run ./cmd/bot d
go run ./cmd/bot w
go run ./cmd/bot q
go run ./cmd/bot y
```

## Documentation

- [設定](docs/configuration.md)
- [運用](docs/operations.md)
- [コード構成](docs/codebase.md)
- [分析ロジック](docs/analytics.md)

## Test

```bash
go test ./...
```

GitHub Actions では `push` / `pull_request` ごとに `go test ./...` を実行します。

## Scheduled Analysis

`.github/workflows/analyze.yaml` で以下の定期実行を設定しています。

- 毎日 09:05 JST: daily
- 毎週月曜 09:15 JST: weekly
- 四半期初日 09:25 JST: quarterly
- 1月1日 09:35 JST: yearly

GitHub Actions で実行する場合は repository secrets に `DISCORD_WEBHOOK_URL` と API key を設定してください。
