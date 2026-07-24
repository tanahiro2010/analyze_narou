# Configuration

設定は `.env`、環境変数、GitHub Actions secrets から読みます。

## Required

### `DISCORD_WEBHOOK_URL`

分析結果を送信する Discord Webhook URL です。

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

### `DENI_API_KEY` または `OPENAI_API_KEY`

AI 分析に使う OpenAI API 互換 API key です。

`DENI_API_KEY` が設定されている場合は `OPENAI_API_KEY` より優先されます。

```env
DENI_API_KEY=deni_xxx
```

## Optional

### `NAROU_URL`

なろう API の base URL です。

Default:

```env
NAROU_URL=https://api.syosetu.com/
```

### `NAROU_USER_AGENT`

なろう API へ送る User-Agent です。

### `NAROU_RANKING_LIMIT`

`novelapi` でランキング相当の作品を取得するときの `lim` です。

Default:

```env
NAROU_RANKING_LIMIT=100
```

### `GENRE_ANALYZE_CONCURRENCY`

ジャンル別解析の並列数です。

Default:

```env
GENRE_ANALYZE_CONCURRENCY=4
```

0 以下の場合は安全のため 1 並列として扱います。

### `OPENAI_BASE_URL`

OpenAI API 互換 API の base URL です。

Default:

```env
OPENAI_BASE_URL=https://api.deniai.app/v1
```

### `OPENAI_MODEL`

AI 分析に使うモデル ID です。

Default:

```env
OPENAI_MODEL=openai/gpt-5.2
```

### `DISCORD_TIMEOUT`

Discord Webhook 送信の timeout です。Go の duration 形式で指定します。

Default:

```env
DISCORD_TIMEOUT=10s
```

## Example `.env`

```env
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
DENI_API_KEY=deni_xxx

NAROU_URL=https://api.syosetu.com/
NAROU_RANKING_LIMIT=100
GENRE_ANALYZE_CONCURRENCY=4

OPENAI_BASE_URL=https://api.deniai.app/v1
OPENAI_MODEL=openai/gpt-5.2
DISCORD_TIMEOUT=10s
```

## Discord Server

この bot のログを送り続けるサーバーは承認制です。

Discord 招待リンク: https://discord.gg/StYV8QMWPp

Webhook はサーバー内の投稿先チャンネルで作成してください。
