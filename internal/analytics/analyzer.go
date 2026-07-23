package analytics

import (
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/sashabaranov/go-openai"
)

type Analyzer struct {
	OpenAIClient ChatClient
}

type ChatClient interface {
	Chat(prompts []openai.ChatCompletionMessage) ([]openai.ChatCompletionResponse, error)
}

func NewAnalyzer(openAIClient ChatClient) *Analyzer {
	return &Analyzer{
		OpenAIClient: openAIClient,
	}
}

func (a *Analyzer) GenreAnalyze(ctx []narou.Novel) (GenreAnalyzeResult, error) {
	fmt.Printf("Analyzing %d novels...\n", len(ctx))
	result := analyzeNovels(ctx)
	a.enrichGenreWithAI(&result)
	return result, nil
}

func (a *Analyzer) AllAnalyze(ctx []GenreAnalyzeResult) (AllAnalyzeResult, error) {
	result := AllAnalyzeResult{
		GenreResultCount: len(ctx),
	}

	tagCounts := map[string]int{}
	var genreSummaries []GenreSummary
	var totalBookmarks int
	var totalEvaluators int
	var totalEvaluationPoints int
	var totalGlobalPoint int
	var totalLength int
	var totalEpisodeCount int
	var minLength int
	var maxLength int
	var topGlobalPoint []NovelDigest
	var serialCount int
	var completedSerialCount int
	var ongoingSerialCount int
	var shortCount int
	var stoppedCount int

	for _, genreResult := range ctx {
		result.NovelCount += genreResult.NovelCount
		totalBookmarks += genreResult.BookmarkAnalysis.TotalBookmarks
		totalEvaluators += genreResult.EvaluationAnalysis.TotalEvaluators
		totalEvaluationPoints += genreResult.EvaluationAnalysis.TotalEvaluationPoints
		totalGlobalPoint += genreResult.PointAnalysis.TotalGlobalPoint
		totalLength += genreResult.LengthAnalysis.TotalLength
		totalEpisodeCount += genreResult.LengthAnalysis.TotalEpisodeCount
		if genreResult.NovelCount > 0 && (minLength == 0 || genreResult.LengthAnalysis.MinLength < minLength) {
			minLength = genreResult.LengthAnalysis.MinLength
		}
		if genreResult.LengthAnalysis.MaxLength > maxLength {
			maxLength = genreResult.LengthAnalysis.MaxLength
		}
		serialCount += genreResult.SerializationAnalysis.SerialCount
		completedSerialCount += genreResult.SerializationAnalysis.CompletedSerialCount
		ongoingSerialCount += genreResult.SerializationAnalysis.OngoingSerialCount
		shortCount += genreResult.SerializationAnalysis.ShortCount
		stoppedCount += genreResult.SerializationAnalysis.StoppedCount
		topGlobalPoint = append(topGlobalPoint, genreResult.PointAnalysis.TopGlobalPoint...)

		for _, tag := range genreResult.TagDistribution {
			tagCounts[tag.Tag] += tag.Count
		}

		for _, genre := range genreResult.GenreDistribution {
			genreSummaries = append(genreSummaries, GenreSummary{
				Genre:               genre.Genre,
				NovelCount:          genre.Count,
				TopTags:             topTagsForGenre(genreResult.TagDistributionByGenre, genre.Genre),
				AverageBookmarkRate: genreResult.BookmarkAnalysis.BookmarkToEvaluatorRate,
				AverageRating:       genreResult.EvaluationAnalysis.AverageRatingPerEvaluator,
				AverageLength:       genreResult.LengthAnalysis.AverageLength,
				AverageGlobalPoint:  genreResult.PointAnalysis.AverageGlobalPoint,
			})
		}

	}

	result.TagDistribution = sortedTagCounts(tagCounts, result.NovelCount)
	result.BookmarkAnalysis = buildBookmarkAnalysis(result.NovelCount, totalBookmarks, totalEvaluators, totalGlobalPoint)
	result.EvaluationAnalysis = buildEvaluationAnalysis(result.NovelCount, totalEvaluationPoints, totalEvaluators)
	result.LengthAnalysis = LengthAnalysis{
		TotalLength:         totalLength,
		AverageLength:       average(totalLength, result.NovelCount),
		MinLength:           minLength,
		MaxLength:           maxLength,
		TotalEpisodeCount:   totalEpisodeCount,
		AverageEpisodeCount: average(totalEpisodeCount, result.NovelCount),
	}
	result.PointAnalysis = PointAnalysis{
		TotalGlobalPoint:   totalGlobalPoint,
		AverageGlobalPoint: average(totalGlobalPoint, result.NovelCount),
		TopGlobalPoint:     topNovelDigests(topGlobalPoint, 5),
	}
	result.SerializationAnalysis = buildSerializationAnalysis(
		result.NovelCount,
		shortCount,
		serialCount,
		completedSerialCount,
		ongoingSerialCount,
		stoppedCount,
	)
	result.GenreSummaries = sortedGenreSummaries(genreSummaries)
	result.WritingHints = buildWritingHints(result)
	a.enrichAllWithAI(&result)

	return result, nil
}

var _ ChatClient = (*gpt.OpenAIClient)(nil)

func (a *Analyzer) enrichGenreWithAI(result *GenreAnalyzeResult) {
	if result.NovelCount == 0 {
		return
	}

	insight, err := a.askAI(buildGenreAIPrompt(*result))
	if err != nil {
		result.AIInsight = AIInsight{UnavailableReason: err.Error()}
		return
	}

	result.AIInsight = insight
}

func (a *Analyzer) enrichAllWithAI(result *AllAnalyzeResult) {
	if result.NovelCount == 0 {
		return
	}

	insight, err := a.askAI(buildAllAIPrompt(*result))
	if err != nil {
		result.AIInsight = AIInsight{UnavailableReason: err.Error()}
		return
	}

	result.AIInsight = insight
}

func (a *Analyzer) askAI(userPrompt string) (AIInsight, error) {
	if a == nil || a.OpenAIClient == nil {
		return AIInsight{}, fmt.Errorf("AI client is not configured")
	}

	responses, err := a.OpenAIClient.Chat([]openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: strings.Join([]string{
				"あなたは小説投稿サイトのランキング分析者です。",
				"与えられた集計値と作品サンプルだけを根拠に、執筆に使える示唆を日本語で返してください。",
				"過度な断定を避け、ランキング上位作品の共通点、紹介文の書き方、タグ設計、読者反応の読みを分けてください。",
				"JSON以外の文章は返さないでください。",
			}, "\n"),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	})
	if err != nil {
		return AIInsight{}, fmt.Errorf("AI analysis failed: %w", err)
	}

	content := firstChatContent(responses)
	if content == "" {
		return AIInsight{}, fmt.Errorf("AI analysis returned empty response")
	}

	insight, err := parseAIInsight(content)
	if err != nil {
		insight.Raw = content
		insight.UnavailableReason = err.Error()
		return insight, nil
	}

	return insight, nil
}

func buildGenreAIPrompt(result GenreAnalyzeResult) string {
	payload := map[string]any{
		"analysis_scope": "genre_ranking",
		"required_output_json_schema": map[string]any{
			"summary":         "ランキング全体の短い要約",
			"title_and_story": "タイトルと小説紹介の傾向。どんなタイトルで、紹介文がどこまで書いているか",
			"tag_and_genre":   "ジャンルごとのタグ分布から見える読者期待",
			"reader_signal":   "ブックマーク率、平均評価、平均文字数などから読む読者反応",
			"writing_advice":  []string{"執筆に使える具体的な示唆を3から5個"},
		},
		"metrics": map[string]any{
			"novel_count":                result.NovelCount,
			"big_genre_distribution":     result.BigGenreDistribution,
			"genre_distribution":         result.GenreDistribution,
			"title_analysis":             result.TitleStoryAnalysis.Title,
			"story_analysis":             result.TitleStoryAnalysis.Story,
			"top_tags":                   limitTags(result.TagDistribution, 15),
			"tag_distribution_by_genre":  result.TagDistributionByGenre,
			"bookmark_analysis":          result.BookmarkAnalysis,
			"evaluation_analysis":        result.EvaluationAnalysis,
			"length_analysis":            result.LengthAnalysis,
			"point_analysis":             result.PointAnalysis,
			"serialization_analysis":     result.SerializationAnalysis,
			"dialogue_analysis":          result.DialogueAnalysis,
			"representative_title_story": result.TitleStoryAnalysis.RepresentativeWork,
			"average_rating_explanation": "AverageRatingPerEvaluator = 評価点合計 / 評価者数",
			"bookmark_rate_explanation":  "BookmarkToEvaluatorRate = ブックマーク数 / 評価者数",
			"bookmark_share_explanation": "BookmarkPointShare = ブックマーク由来ポイント推定 / 総合ポイント",
		},
	}

	return mustMarshalPrompt(payload)
}

func buildAllAIPrompt(result AllAnalyzeResult) string {
	payload := map[string]any{
		"analysis_scope": "all_rankings",
		"required_output_json_schema": map[string]any{
			"summary":         "全ジャンル横断の短い要約",
			"title_and_story": "全体として強いタイトルと紹介文の見せ方",
			"tag_and_genre":   "ジャンル別タグ分布と狙い目",
			"reader_signal":   "ブックマーク率、平均評価、平均文字数などから読む読者反応",
			"writing_advice":  []string{"新作企画や既存作改善に使える具体的な示唆を3から5個"},
		},
		"metrics": map[string]any{
			"genre_result_count":          result.GenreResultCount,
			"novel_count":                 result.NovelCount,
			"top_tags":                    limitTags(result.TagDistribution, 20),
			"bookmark_analysis":           result.BookmarkAnalysis,
			"evaluation_analysis":         result.EvaluationAnalysis,
			"length_analysis":             result.LengthAnalysis,
			"point_analysis":              result.PointAnalysis,
			"serialization_analysis":      result.SerializationAnalysis,
			"genre_summaries":             result.GenreSummaries,
			"deterministic_writing_hints": result.WritingHints,
		},
	}

	return mustMarshalPrompt(payload)
}

func firstChatContent(responses []openai.ChatCompletionResponse) string {
	for _, response := range responses {
		for _, choice := range response.Choices {
			content := strings.TrimSpace(choice.Message.Content)
			if content != "" {
				return content
			}
		}
	}
	return ""
}

func parseAIInsight(content string) (AIInsight, error) {
	var payload struct {
		Summary       string   `json:"summary"`
		TitleAndStory string   `json:"title_and_story"`
		TagAndGenre   string   `json:"tag_and_genre"`
		ReaderSignal  string   `json:"reader_signal"`
		WritingAdvice []string `json:"writing_advice"`
	}

	if err := json.Unmarshal([]byte(stripJSONFence(content)), &payload); err != nil {
		return AIInsight{}, fmt.Errorf("parse AI insight JSON: %w", err)
	}

	return AIInsight{
		Summary:       payload.Summary,
		TitleAndStory: payload.TitleAndStory,
		TagAndGenre:   payload.TagAndGenre,
		ReaderSignal:  payload.ReaderSignal,
		WritingAdvice: payload.WritingAdvice,
	}, nil
}

func stripJSONFence(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

func mustMarshalPrompt(payload any) string {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", payload)
	}
	return string(data)
}

func limitTags(tags []TagCount, limit int) []TagCount {
	if len(tags) > limit {
		return tags[:limit]
	}
	return tags
}

func analyzeNovels(novels []narou.Novel) GenreAnalyzeResult {
	novelCount := len(novels)
	result := GenreAnalyzeResult{
		NovelCount: novelCount,
	}

	if novelCount == 0 {
		return result
	}

	bigGenreCounts := map[narou.BigGenre]int{}
	genreCounts := map[narou.Genre]int{}
	tagCounts := map[string]int{}
	genreTagCounts := map[narou.Genre]map[string]int{}
	termCounts := map[string]int{}
	var totalTitleLength int
	var minTitleLength int
	var maxTitleLength int
	var longTitleCount int
	var questionOrExclamationCount int
	var bracketTitleCount int
	var totalStoryLength int
	var minStoryLength int
	var maxStoryLength int
	var depth StoryDepthDistribution
	var totalBookmarks int
	var totalEvaluators int
	var totalEvaluationPoints int
	var totalLength int
	var totalEpisodeCount int
	var totalGlobalPoint int
	var totalDialogueRate int
	var lengths []int
	var digests []NovelDigest
	var shortCount int
	var serialCount int
	var completedSerialCount int
	var ongoingSerialCount int
	var stoppedCount int

	for index, novel := range novels {
		bigGenreCounts[novel.BigGenre]++
		genreCounts[novel.Genre]++

		titleLength := runeLength(novel.Title)
		storyLength := runeLength(novel.Story)
		if index == 0 || titleLength < minTitleLength {
			minTitleLength = titleLength
		}
		if titleLength > maxTitleLength {
			maxTitleLength = titleLength
		}
		if index == 0 || storyLength < minStoryLength {
			minStoryLength = storyLength
		}
		if storyLength > maxStoryLength {
			maxStoryLength = storyLength
		}

		totalTitleLength += titleLength
		totalStoryLength += storyLength
		if titleLength >= 30 {
			longTitleCount++
		}
		if strings.ContainsAny(novel.Title, "!?！？") {
			questionOrExclamationCount++
		}
		if strings.ContainsAny(novel.Title, "「」『』【】[]（）()") {
			bracketTitleCount++
		}

		switch classifyStoryDepth(novel.Story) {
		case "ending":
			depth.EndingOrSpoiler++
		case "development":
			depth.DevelopmentIncluded++
		case "goal":
			depth.GoalOrConflict++
		default:
			depth.SetupOnly++
		}

		for _, tag := range splitTags(novel.Keyword) {
			tagCounts[tag]++
			if _, ok := genreTagCounts[novel.Genre]; !ok {
				genreTagCounts[novel.Genre] = map[string]int{}
			}
			genreTagCounts[novel.Genre][tag]++
		}
		for _, term := range extractTerms(novel.Title + " " + novel.Story) {
			termCounts[term]++
		}

		totalBookmarks += novel.FavoriteNovelCount
		totalEvaluators += novel.EvaluatorCount
		totalEvaluationPoints += novel.EvaluationPoint
		totalLength += novel.Length
		totalEpisodeCount += novel.EpisodeCount
		totalGlobalPoint += novel.GlobalPoint
		totalDialogueRate += novel.DialogueRate
		lengths = append(lengths, novel.Length)
		digests = append(digests, NovelDigest{
			NCode:       novel.NCode,
			Title:       novel.Title,
			StoryDigest: truncateRunes(novel.Story, 120),
			GlobalPoint: novel.GlobalPoint,
			Length:      novel.Length,
		})

		if novel.IsShort() {
			shortCount++
		}
		if novel.IsSerial() {
			serialCount++
		}
		if novel.IsCompletedSerial() {
			completedSerialCount++
		}
		if novel.IsOngoingSerial() {
			ongoingSerialCount++
		}
		if novel.IsStopped.Bool() {
			stoppedCount++
		}
	}

	result.BigGenreDistribution = sortedBigGenreCounts(bigGenreCounts, novelCount)
	result.GenreDistribution = sortedGenreCounts(genreCounts, novelCount)
	result.TitleStoryAnalysis = TitleStoryAnalysis{
		Title: TitleAnalysis{
			AverageLength:             average(totalTitleLength, novelCount),
			MinLength:                 minTitleLength,
			MaxLength:                 maxTitleLength,
			LongTitleRate:             average(longTitleCount, novelCount),
			QuestionOrExclamationRate: average(questionOrExclamationCount, novelCount),
			BracketTitleRate:          average(bracketTitleCount, novelCount),
		},
		Story: StoryAnalysis{
			AverageLength:     average(totalStoryLength, novelCount),
			MinLength:         minStoryLength,
			MaxLength:         maxStoryLength,
			DepthDistribution: depth,
			CommonTerms:       sortedTermCounts(termCounts, novelCount, 10),
		},
		RepresentativeWork: topNovelDigests(digests, 5),
	}
	result.TagDistribution = sortedTagCounts(tagCounts, novelCount)
	result.TagDistributionByGenre = sortedGenreTagDistribution(genreTagCounts, genreCounts)
	result.BookmarkAnalysis = buildBookmarkAnalysis(novelCount, totalBookmarks, totalEvaluators, totalGlobalPoint)
	result.EvaluationAnalysis = buildEvaluationAnalysis(novelCount, totalEvaluationPoints, totalEvaluators)
	result.LengthAnalysis = buildLengthAnalysis(novelCount, totalLength, totalEpisodeCount, lengths)
	result.PointAnalysis = PointAnalysis{
		TotalGlobalPoint:   totalGlobalPoint,
		AverageGlobalPoint: average(totalGlobalPoint, novelCount),
		TopGlobalPoint:     topNovelDigests(digests, 5),
	}
	result.SerializationAnalysis = buildSerializationAnalysis(
		novelCount,
		shortCount,
		serialCount,
		completedSerialCount,
		ongoingSerialCount,
		stoppedCount,
	)
	result.DialogueAnalysis = DialogueAnalysis{
		AverageDialogueRate: average(totalDialogueRate, novelCount),
	}

	return result
}

func buildBookmarkAnalysis(novelCount, totalBookmarks, totalEvaluators, totalGlobalPoint int) BookmarkAnalysis {
	return BookmarkAnalysis{
		TotalBookmarks:          totalBookmarks,
		AverageBookmarks:        average(totalBookmarks, novelCount),
		BookmarkToEvaluatorRate: ratio(totalBookmarks, totalEvaluators),
		BookmarkPointShare:      ratio(totalBookmarks*2, totalGlobalPoint),
	}
}

func buildEvaluationAnalysis(novelCount, totalEvaluationPoints, totalEvaluators int) EvaluationAnalysis {
	return EvaluationAnalysis{
		TotalEvaluationPoints:     totalEvaluationPoints,
		TotalEvaluators:           totalEvaluators,
		AverageEvaluationPoint:    average(totalEvaluationPoints, novelCount),
		AverageEvaluatorCount:     average(totalEvaluators, novelCount),
		AverageRatingPerEvaluator: ratio(totalEvaluationPoints, totalEvaluators),
	}
}

func buildLengthAnalysis(novelCount, totalLength, totalEpisodeCount int, lengths []int) LengthAnalysis {
	return LengthAnalysis{
		TotalLength:         totalLength,
		AverageLength:       average(totalLength, novelCount),
		MedianLength:        median(lengths),
		MinLength:           minInt(lengths),
		MaxLength:           maxInt(lengths),
		TotalEpisodeCount:   totalEpisodeCount,
		AverageEpisodeCount: average(totalEpisodeCount, novelCount),
	}
}

func buildSerializationAnalysis(novelCount, shortCount, serialCount, completedSerialCount, ongoingSerialCount, stoppedCount int) SerializationAnalysis {
	return SerializationAnalysis{
		ShortCount:           shortCount,
		SerialCount:          serialCount,
		CompletedSerialCount: completedSerialCount,
		OngoingSerialCount:   ongoingSerialCount,
		StoppedCount:         stoppedCount,
		CompletionRate:       ratio(completedSerialCount, serialCount),
		StoppedRate:          average(stoppedCount, novelCount),
	}
}

func sortedBigGenreCounts(counts map[narou.BigGenre]int, total int) []BigGenreCount {
	results := make([]BigGenreCount, 0, len(counts))
	for genre, count := range counts {
		results = append(results, BigGenreCount{
			BigGenre: genre,
			Count:    count,
			Rate:     average(count, total),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].BigGenre < results[j].BigGenre
		}
		return results[i].Count > results[j].Count
	})
	return results
}

func sortedGenreCounts(counts map[narou.Genre]int, total int) []GenreCount {
	results := make([]GenreCount, 0, len(counts))
	for genre, count := range counts {
		results = append(results, GenreCount{
			Genre: genre,
			Count: count,
			Rate:  average(count, total),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Genre < results[j].Genre
		}
		return results[i].Count > results[j].Count
	})
	return results
}

func sortedTagCounts(counts map[string]int, totalNovels int) []TagCount {
	results := make([]TagCount, 0, len(counts))
	for tag, count := range counts {
		results = append(results, TagCount{
			Tag:   tag,
			Count: count,
			Rate:  average(count, totalNovels),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Tag < results[j].Tag
		}
		return results[i].Count > results[j].Count
	})
	return results
}

func sortedTermCounts(counts map[string]int, totalNovels int, limit int) []TermCount {
	results := make([]TermCount, 0, len(counts))
	for term, count := range counts {
		results = append(results, TermCount{
			Term:  term,
			Count: count,
			Rate:  average(count, totalNovels),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Term < results[j].Term
		}
		return results[i].Count > results[j].Count
	})
	if len(results) > limit {
		return results[:limit]
	}
	return results
}

func sortedGenreTagDistribution(tagCounts map[narou.Genre]map[string]int, genreCounts map[narou.Genre]int) []GenreTagDistribution {
	results := make([]GenreTagDistribution, 0, len(tagCounts))
	for genre, counts := range tagCounts {
		results = append(results, GenreTagDistribution{
			Genre: genre,
			Count: genreCounts[genre],
			Tags:  topTagCounts(counts, genreCounts[genre], 10),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Genre < results[j].Genre
		}
		return results[i].Count > results[j].Count
	})
	return results
}

func topTagCounts(counts map[string]int, totalNovels int, limit int) []TagCount {
	tags := sortedTagCounts(counts, totalNovels)
	if len(tags) > limit {
		return tags[:limit]
	}
	return tags
}

func topTagsForGenre(distribution []GenreTagDistribution, genre narou.Genre) []TagCount {
	for _, item := range distribution {
		if item.Genre == genre {
			return item.Tags
		}
	}
	return nil
}

func sortedGenreSummaries(summaries []GenreSummary) []GenreSummary {
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].NovelCount == summaries[j].NovelCount {
			return summaries[i].Genre < summaries[j].Genre
		}
		return summaries[i].NovelCount > summaries[j].NovelCount
	})
	return summaries
}

func topNovelDigests(novels []NovelDigest, limit int) []NovelDigest {
	sort.Slice(novels, func(i, j int) bool {
		if novels[i].GlobalPoint == novels[j].GlobalPoint {
			return novels[i].NCode < novels[j].NCode
		}
		return novels[i].GlobalPoint > novels[j].GlobalPoint
	})
	if len(novels) > limit {
		return novels[:limit]
	}
	return novels
}

func buildWritingHints(result AllAnalyzeResult) []string {
	if result.NovelCount == 0 {
		return nil
	}

	hints := []string{}
	if len(result.TagDistribution) > 0 {
		hints = append(hints, fmt.Sprintf("上位タグは「%s」。同ジャンルで読者が期待している要素として優先確認する。", result.TagDistribution[0].Tag))
	}
	if result.BookmarkAnalysis.BookmarkPointShare >= 0.5 {
		hints = append(hints, "総合ポイントに占めるブックマーク寄与が高い。序盤で継続読書したくなる目的や関係性を明確にする。")
	}
	if result.SerializationAnalysis.CompletionRate >= 0.5 {
		hints = append(hints, "完結済み連載の比率が高い。完結保証や到達点を紹介文で示すと比較されやすい。")
	}
	if result.LengthAnalysis.AverageLength > 0 {
		hints = append(hints, fmt.Sprintf("平均文字数は%.0f字。上位作品の分量感に合わせて、序盤の更新量と章立てを設計する。", result.LengthAnalysis.AverageLength))
	}
	return hints
}

func splitTags(value string) []string {
	fields := strings.Fields(value)
	tags := make([]string, 0, len(fields))
	for _, field := range fields {
		tag := strings.Trim(field, " \t\r\n、,")
		if tag == "" {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

func extractTerms(value string) []string {
	splitter := func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || strings.ContainsRune("、。，．・「」『』【】（）()［］[]!?！？…ー〜~", r)
	}

	parts := strings.FieldsFunc(value, splitter)
	terms := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if runeLength(part) < 2 {
			continue
		}
		terms = append(terms, part)
	}
	return terms
}

func classifyStoryDepth(story string) string {
	switch {
	case containsAny(story, []string{"結末", "最後", "ラスト", "真相", "ネタバレ", "正体"}):
		return "ending"
	case containsAny(story, []string{"やがて", "しかし", "だが", "ところが", "一方", "そして", "始まる", "展開"}):
		return "development"
	case containsAny(story, []string{"目指", "ために", "守る", "戦", "挑む", "救", "探", "復讐", "解決", "巻き込"}):
		return "goal"
	default:
		return "setup"
	}
}

func containsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func average(total, count int) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func median(values []int) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]int(nil), values...)
	sort.Ints(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return float64(sorted[mid])
	}
	return float64(sorted[mid-1]+sorted[mid]) / 2
}

func minInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	minValue := values[0]
	for _, value := range values[1:] {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func maxInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	maxValue := values[0]
	for _, value := range values[1:] {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func runeLength(value string) int {
	return len([]rune(value))
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "..."
}

func (r GenreAnalyzeResult) String() string {
	if r.NovelCount == 0 {
		return "作品数: 0"
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "作品数: %d\n", r.NovelCount)
	fmt.Fprintf(&builder, "タイトル: 平均%.1f字 / 長文率%.1f%% / 疑問・感嘆符率%.1f%% / 括弧入り率%.1f%%\n",
		r.TitleStoryAnalysis.Title.AverageLength,
		r.TitleStoryAnalysis.Title.LongTitleRate*100,
		r.TitleStoryAnalysis.Title.QuestionOrExclamationRate*100,
		r.TitleStoryAnalysis.Title.BracketTitleRate*100,
	)
	fmt.Fprintf(&builder, "紹介文: 平均%.1f字 / 導入のみ%d / 目的・対立まで%d / 展開まで%d / 結末示唆%d\n",
		r.TitleStoryAnalysis.Story.AverageLength,
		r.TitleStoryAnalysis.Story.DepthDistribution.SetupOnly,
		r.TitleStoryAnalysis.Story.DepthDistribution.GoalOrConflict,
		r.TitleStoryAnalysis.Story.DepthDistribution.DevelopmentIncluded,
		r.TitleStoryAnalysis.Story.DepthDistribution.EndingOrSpoiler,
	)
	fmt.Fprintf(&builder, "ブックマーク: 平均%.1f件 / 評価者比%.2f / ポイント寄与%.1f%%\n",
		r.BookmarkAnalysis.AverageBookmarks,
		r.BookmarkAnalysis.BookmarkToEvaluatorRate,
		r.BookmarkAnalysis.BookmarkPointShare*100,
	)
	fmt.Fprintf(&builder, "評価: 平均評価点%.1f / 評価者あたり%.2f / 平均文字数%.0f字\n",
		r.EvaluationAnalysis.AverageEvaluationPoint,
		r.EvaluationAnalysis.AverageRatingPerEvaluator,
		r.LengthAnalysis.AverageLength,
	)
	if len(r.TagDistribution) > 0 {
		fmt.Fprintf(&builder, "上位タグ: %s\n", formatTags(r.TagDistribution, 10))
	}
	appendAIInsight(&builder, r.AIInsight)
	return strings.TrimSpace(builder.String())
}

func (r AllAnalyzeResult) String() string {
	if r.NovelCount == 0 {
		return "作品数: 0"
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "総作品数: %d / 分析グループ数: %d\n", r.NovelCount, r.GenreResultCount)
	fmt.Fprintf(&builder, "ブックマーク: 平均%.1f件 / 評価者比%.2f / ポイント寄与%.1f%%\n",
		r.BookmarkAnalysis.AverageBookmarks,
		r.BookmarkAnalysis.BookmarkToEvaluatorRate,
		r.BookmarkAnalysis.BookmarkPointShare*100,
	)
	fmt.Fprintf(&builder, "評価: 評価者あたり%.2f点 / 平均文字数%.0f字 / 平均総合ポイント%.1f\n",
		r.EvaluationAnalysis.AverageRatingPerEvaluator,
		r.LengthAnalysis.AverageLength,
		r.PointAnalysis.AverageGlobalPoint,
	)
	if len(r.TagDistribution) > 0 {
		fmt.Fprintf(&builder, "全体上位タグ: %s\n", formatTags(r.TagDistribution, 10))
	}
	if len(r.WritingHints) > 0 {
		fmt.Fprintf(&builder, "執筆ヒント: %s\n", strings.Join(r.WritingHints, " / "))
	}
	appendAIInsight(&builder, r.AIInsight)
	return strings.TrimSpace(builder.String())
}

func appendAIInsight(builder *strings.Builder, insight AIInsight) {
	if insight.Summary != "" {
		fmt.Fprintf(builder, "AI要約: %s\n", insight.Summary)
	}
	if insight.TitleAndStory != "" {
		fmt.Fprintf(builder, "AIタイトル・紹介分析: %s\n", insight.TitleAndStory)
	}
	if insight.TagAndGenre != "" {
		fmt.Fprintf(builder, "AIタグ・ジャンル分析: %s\n", insight.TagAndGenre)
	}
	if insight.ReaderSignal != "" {
		fmt.Fprintf(builder, "AI読者反応分析: %s\n", insight.ReaderSignal)
	}
	if len(insight.WritingAdvice) > 0 {
		fmt.Fprintf(builder, "AI執筆アドバイス: %s\n", strings.Join(insight.WritingAdvice, " / "))
	}
	if insight.UnavailableReason != "" {
		fmt.Fprintf(builder, "AI分析: 利用不可 (%s)\n", insight.UnavailableReason)
	}
}

func formatTags(tags []TagCount, limit int) string {
	if len(tags) > limit {
		tags = tags[:limit]
	}
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, fmt.Sprintf("%s(%d)", tag.Tag, tag.Count))
	}
	return strings.Join(parts, ", ")
}
