# Analytics Internals

このドキュメントは `internal/analytics` の詳細です。

## Entry Points

### `Analyzer.GenreAnalyze`

```go
func (a *Analyzer) GenreAnalyze(ctx []narou.Novel, targetGenreName ...string) (GenreAnalyzeResult, error)
```

1つのランキング群、つまり1ジャンル分の `[]narou.Novel` を分析します。

処理:

1. `analyzeNovels` で deterministic な集計値を作る
2. `targetGenreName` が渡されていれば `TargetGenreName` に保存する
3. `enrichGenreWithAI` で AI 要約を付与する

AI client が nil の場合、集計値はそのまま返り、AI 欄には利用不可理由が入ります。

### `Analyzer.AllAnalyze`

```go
func (a *Analyzer) AllAnalyze(ctx []GenreAnalyzeResult) (AllAnalyzeResult, error)
```

ジャンル別分析結果を横断集計します。

処理:

1. 作品数、ブックマーク、評価、文字数、ポイントなどを合算する
2. タグ分布を再集計する
3. ジャンル別 summary を作る
4. deterministic な writing hints を作る
5. `enrichAllWithAI` で全体 AI 要約を付与する

## Genre Analysis Metrics

`analyzeNovels` が計算します。

### Genre Distribution

- `BigGenreDistribution`
- `GenreDistribution`
- `TagDistributionByGenre`

なろうの `biggenre` / `genre` コードごとの件数と比率を作ります。

AI prompt ではコードだけだと読みづらいため、名前付きの payload も追加します。

- `big_genre_distribution_with_names`
- `genre_distribution_with_names`
- `tag_distribution_by_genre_with_names`

### Title Analysis

`TitleAnalysis`:

- `AverageLength`
- `MinLength`
- `MaxLength`
- `LongTitleRate`
- `QuestionOrExclamationRate`
- `BracketTitleRate`

現在の判定:

- 長タイトル: 30文字以上
- 疑問・感嘆符: `!?！？` を含む
- 括弧入り: `「」『』【】[]（）()` などを含む

### Story Analysis

`StoryAnalysis`:

- `AverageLength`
- `MinLength`
- `MaxLength`
- `DepthDistribution`
- `CommonTerms`

`classifyStoryDepth` は紹介文の深さを簡易分類します。

| Result | 判定語の例 |
| --- | --- |
| `ending` | 結末、最後、ラスト、真相、ネタバレ、正体 |
| `development` | やがて、しかし、だが、ところが、一方、そして、始まる、展開 |
| `goal` | 目指、ために、守る、戦、挑む、救、探、復讐、解決、巻き込 |
| `setup` | 上記以外 |

これは厳密な自然言語処理ではなく、ランキング傾向を見るための lightweight heuristic です。

### Tags

`splitTags` は `Novel.Keyword` を空白で分割し、読点や comma を trim します。

`TagDistribution` は全体の上位タグです。

`TagDistributionByGenre` はジャンル別の上位タグです。

### Bookmark Analysis

`BookmarkAnalysis`:

- `TotalBookmarks`
- `AverageBookmarks`
- `BookmarkToEvaluatorRate`
- `BookmarkPointShare`

計算:

```text
AverageBookmarks = TotalBookmarks / NovelCount
BookmarkToEvaluatorRate = FavoriteNovelCount合計 / EvaluatorCount合計
BookmarkPointShare = FavoriteNovelCount合計 * 2 / GlobalPoint合計
```

`BookmarkPointShare` は、なろうの総合ポイントのうちブックマーク由来と推定される割合です。

### Evaluation Analysis

`EvaluationAnalysis`:

- `TotalEvaluationPoints`
- `TotalEvaluators`
- `AverageEvaluationPoint`
- `AverageEvaluatorCount`
- `AverageRatingPerEvaluator`

計算:

```text
AverageRatingPerEvaluator = EvaluationPoint合計 / EvaluatorCount合計
```

### Length Analysis

`LengthAnalysis`:

- `TotalLength`
- `AverageLength`
- `MedianLength`
- `MinLength`
- `MaxLength`
- `TotalEpisodeCount`
- `AverageEpisodeCount`

`MedianLength` は lengths を sort して中央値を取ります。

### Point Analysis

`PointAnalysis`:

- `TotalGlobalPoint`
- `AverageGlobalPoint`
- `TopGlobalPoint`

`TopGlobalPoint` は `NovelDigest` の上位5件です。

### Serialization Analysis

`SerializationAnalysis`:

- `ShortCount`
- `SerialCount`
- `CompletedSerialCount`
- `OngoingSerialCount`
- `StoppedCount`
- `CompletionRate`
- `StoppedRate`

`NovelType` と `EndStatus` から判定します。

### Dialogue Analysis

`DialogueAnalysis`:

- `AverageDialogueRate`

なろう API の `kaiwaritu` を平均します。

## AI Prompt

AI は deterministic metrics の後段です。AI が失敗しても集計自体は失敗扱いにしません。

### Genre Prompt

`buildGenreAIPrompt` が JSON prompt を作ります。

主な payload:

- `analysis_scope: "genre_ranking"`
- `target_genre_name`
- `required_output_json_schema`
- `metrics`

`metrics` にはタイトル分析、紹介文分析、タグ分布、評価、ブックマーク、文字数、代表作品などが入ります。

ジャンル名を AI に明示するため、コードのみの分布とは別に名前付きデータも入れています。

### All Prompt

`buildAllAIPrompt` が全体分析用の JSON prompt を作ります。

主な payload:

- `analysis_scope: "all_rankings"`
- `required_output_json_schema`
- `metrics.genre_summaries`
- `metrics.genre_summaries_with_names`
- `metrics.deterministic_writing_hints`

### Expected AI Response

AI は JSON だけを返す前提です。

```json
{
  "summary": "...",
  "title_and_story": "...",
  "tag_and_genre": "...",
  "reader_signal": "...",
  "writing_advice": ["...", "..."]
}
```

`parseAIInsight` がこの JSON を `AIInsight` に変換します。

コードフェンス付きの JSON も `stripJSONFence` で許容します。

## Discord Formatting

`internal/logger/webhook.go` が `GenreAnalyzeResult` / `AllAnalyzeResult` を Discord embed に変換します。

ジャンル別 embed:

- title: `ジャンル別ランキング分析` または `ジャンル別ランキング分析: {TargetGenreName}`
- description: AI 要約
- fields:
  - 概要
  - タイトル
  - 紹介文
  - ブックマーク
  - 評価・文字数
  - 上位タグ
  - AI タイトル・紹介分析
  - AI タグ・ジャンル分析
  - AI 読者反応分析
  - AI 執筆アドバイス

全体 embed:

- title: `全体ランキング分析`
- description: AI 要約
- fields:
  - 概要
  - ブックマーク
  - 評価・文字数
  - 全体上位タグ
  - 執筆ヒント
  - AI fields

長文 field は `discordEmbedFieldLimit` で切り詰めます。

## Deterministic Hints

`buildWritingHints` は AI なしでも出せる簡易助言です。

現在の条件:

- 上位タグがある
- ブックマーク寄与が 0.5 以上
- 完結率が 0.5 以上
- 平均文字数が 0 より大きい

これは `AllAnalyzeResult.WritingHints` に入ります。

## Extension Notes

### AI prompt に項目を追加する

1. `GenreAnalyzeResult` または `AllAnalyzeResult` に指標を追加
2. `buildGenreAIPrompt` または `buildAllAIPrompt` の `metrics` に追加
3. AI response schema の説明が必要なら `required_output_json_schema` も更新
4. prompt 内容を検証するテストを追加

### Discord embed に項目を追加する

1. `genreAnalyzeEmbed` または `allAnalyzeEmbed` に field を追加
2. 長文になる値は `embedField` を通す
3. `webhook_test.go` で payload を確認

### 新しい分析カテゴリを追加する

1. `model.go` に struct を追加
2. `analyzeNovels` または `AllAnalyze` で算出
3. `String` / AI prompt / Discord embed / tests を更新
