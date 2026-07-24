package logger

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const discordMessageLimit = 1900
const discordEmbedFieldLimit = 1000

const (
	discordGenreColor = 0x4F8EF7
	discordAllColor   = 0x57C785
)

type WebhookLogger struct {
	DiscordClient discord.DiscordClient
}

func NewWebhookLogger(discordClient discord.DiscordClient) *WebhookLogger {
	return &WebhookLogger{
		DiscordClient: discordClient,
	}
}

func (w *WebhookLogger) Log(message string) error {
	for _, content := range splitDiscordMessage(message, discordMessageLimit) {
		if err := w.send(discord.WebhookMessage{
			Username: "Narou Analyzer",
			Content:  content,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (w *WebhookLogger) H1(title string) error {
	return w.Log("# " + title)
}

func (w *WebhookLogger) GenreAnalyzeResult(ctx analytics.GenreAnalyzeResult) error {
	return w.send(discord.WebhookMessage{
		Username: "Narou Analyzer",
		Content:  "ジャンル別ランキング分析",
		Embeds:   []discord.WebhookEmbed{genreAnalyzeEmbed(ctx)},
	})
}

func (w *WebhookLogger) AllAnalyzeResult(ctx analytics.AllAnalyzeResult) error {
	return w.send(discord.WebhookMessage{
		Username: "Narou Analyzer",
		Content:  "全体ランキング分析",
		Embeds:   []discord.WebhookEmbed{allAnalyzeEmbed(ctx)},
	})
}

func (w *WebhookLogger) send(message discord.WebhookMessage) error {
	resp, err := w.DiscordClient.SendMessage(message)
	if err != nil {
		return fmt.Errorf("send discord webhook message: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("send discord webhook message: empty response")
	}

	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"send discord webhook message: status %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}
	if readErr != nil {
		return fmt.Errorf("read discord webhook response: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close discord webhook response: %w", closeErr)
	}

	return nil
}

func genreAnalyzeEmbed(ctx analytics.GenreAnalyzeResult) discord.WebhookEmbed {
	title := genreAnalyzeTitle(ctx.TargetGenreName)
	if ctx.NovelCount == 0 {
		return discord.WebhookEmbed{
			Title:       title,
			Description: "作品数: 0",
			Color:       discordGenreColor,
		}
	}

	fields := []discord.WebhookEmbedField{
		embedField("概要", fmt.Sprintf("作品数: **%d**", ctx.NovelCount), true),
		embedField("タイトル", fmt.Sprintf(
			"平均 %.1f字\n長文率 %.1f%%\n疑問・感嘆符率 %.1f%%\n括弧入り率 %.1f%%",
			ctx.TitleStoryAnalysis.Title.AverageLength,
			ctx.TitleStoryAnalysis.Title.LongTitleRate*100,
			ctx.TitleStoryAnalysis.Title.QuestionOrExclamationRate*100,
			ctx.TitleStoryAnalysis.Title.BracketTitleRate*100,
		), true),
		embedField("紹介文", fmt.Sprintf(
			"平均 %.1f字\n導入のみ %d\n目的・対立まで %d\n展開まで %d\n結末示唆 %d",
			ctx.TitleStoryAnalysis.Story.AverageLength,
			ctx.TitleStoryAnalysis.Story.DepthDistribution.SetupOnly,
			ctx.TitleStoryAnalysis.Story.DepthDistribution.GoalOrConflict,
			ctx.TitleStoryAnalysis.Story.DepthDistribution.DevelopmentIncluded,
			ctx.TitleStoryAnalysis.Story.DepthDistribution.EndingOrSpoiler,
		), true),
		embedField("ブックマーク", fmt.Sprintf(
			"平均 %.1f件\n評価者比 %.2f\nポイント寄与 %.1f%%",
			ctx.BookmarkAnalysis.AverageBookmarks,
			ctx.BookmarkAnalysis.BookmarkToEvaluatorRate,
			ctx.BookmarkAnalysis.BookmarkPointShare*100,
		), true),
		embedField("評価・文字数", fmt.Sprintf(
			"平均評価点 %.1f\n評価者あたり %.2f\n平均文字数 %.0f字",
			ctx.EvaluationAnalysis.AverageEvaluationPoint,
			ctx.EvaluationAnalysis.AverageRatingPerEvaluator,
			ctx.LengthAnalysis.AverageLength,
		), true),
	}
	if len(ctx.TagDistribution) > 0 {
		fields = append(fields, embedField("上位タグ", formatDiscordTags(ctx.TagDistribution, 10), false))
	}
	fields = appendAIInsightFields(fields, ctx.AIInsight)

	return discord.WebhookEmbed{
		Title:       title,
		Description: aiSummaryDescription(ctx.AIInsight),
		Color:       discordGenreColor,
		Fields:      fields,
	}
}

func genreAnalyzeTitle(targetGenreName string) string {
	if targetGenreName == "" {
		return "ジャンル別ランキング分析"
	}

	return "ジャンル別ランキング分析: " + targetGenreName
}

func allAnalyzeEmbed(ctx analytics.AllAnalyzeResult) discord.WebhookEmbed {
	if ctx.NovelCount == 0 {
		return discord.WebhookEmbed{
			Title:       "全体ランキング分析",
			Description: "作品数: 0",
			Color:       discordAllColor,
		}
	}

	fields := []discord.WebhookEmbedField{
		embedField("概要", fmt.Sprintf(
			"総作品数: **%d**\n分析グループ数: **%d**",
			ctx.NovelCount,
			ctx.GenreResultCount,
		), true),
		embedField("ブックマーク", fmt.Sprintf(
			"平均 %.1f件\n評価者比 %.2f\nポイント寄与 %.1f%%",
			ctx.BookmarkAnalysis.AverageBookmarks,
			ctx.BookmarkAnalysis.BookmarkToEvaluatorRate,
			ctx.BookmarkAnalysis.BookmarkPointShare*100,
		), true),
		embedField("評価・文字数", fmt.Sprintf(
			"評価者あたり %.2f点\n平均文字数 %.0f字\n平均総合ポイント %.1f",
			ctx.EvaluationAnalysis.AverageRatingPerEvaluator,
			ctx.LengthAnalysis.AverageLength,
			ctx.PointAnalysis.AverageGlobalPoint,
		), true),
	}
	if len(ctx.TagDistribution) > 0 {
		fields = append(fields, embedField("全体上位タグ", formatDiscordTags(ctx.TagDistribution, 10), false))
	}
	if len(ctx.WritingHints) > 0 {
		fields = append(fields, embedField("執筆ヒント", strings.Join(ctx.WritingHints, "\n"), false))
	}
	fields = appendAIInsightFields(fields, ctx.AIInsight)

	return discord.WebhookEmbed{
		Title:       "全体ランキング分析",
		Description: aiSummaryDescription(ctx.AIInsight),
		Color:       discordAllColor,
		Fields:      fields,
	}
}

func appendAIInsightFields(fields []discord.WebhookEmbedField, insight analytics.AIInsight) []discord.WebhookEmbedField {
	if insight.TitleAndStory != "" {
		fields = append(fields, embedField("AI タイトル・紹介分析", insight.TitleAndStory, false))
	}
	if insight.TagAndGenre != "" {
		fields = append(fields, embedField("AI タグ・ジャンル分析", insight.TagAndGenre, false))
	}
	if insight.ReaderSignal != "" {
		fields = append(fields, embedField("AI 読者反応分析", insight.ReaderSignal, false))
	}
	if len(insight.WritingAdvice) > 0 {
		fields = append(fields, embedField("AI 執筆アドバイス", strings.Join(insight.WritingAdvice, "\n"), false))
	}
	if insight.UnavailableReason != "" {
		fields = append(fields, embedField("AI 分析", "利用不可: "+insight.UnavailableReason, false))
	}

	return fields
}

func aiSummaryDescription(insight analytics.AIInsight) string {
	if insight.Summary == "" {
		return ""
	}

	return truncateDiscordText("**AI要約**\n"+quoteLines(insight.Summary), discordEmbedFieldLimit)
}

func embedField(name string, value string, inline bool) discord.WebhookEmbedField {
	return discord.WebhookEmbedField{
		Name:   name,
		Value:  truncateDiscordText(value, discordEmbedFieldLimit),
		Inline: inline,
	}
}

func formatDiscordTags(tags []analytics.TagCount, limit int) string {
	if len(tags) > limit {
		tags = tags[:limit]
	}

	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, fmt.Sprintf("`%s` %d件", tag.Tag, tag.Count))
	}

	return strings.Join(parts, "\n")
}

func quoteLines(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	for i, line := range lines {
		lines[i] = "> " + strings.TrimSpace(line)
	}

	return strings.Join(lines, "\n")
}

func truncateDiscordText(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return strings.TrimSpace(string(runes[:limit-3])) + "..."
}

func splitDiscordMessage(message string, limit int) []string {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil
	}

	runes := []rune(message)
	if len(runes) <= limit {
		return []string{message}
	}

	var chunks []string
	for len(runes) > 0 {
		end := limit
		if len(runes) < end {
			end = len(runes)
		}

		splitAt := end
		for i := end - 1; i > 0; i-- {
			if runes[i] == '\n' {
				splitAt = i
				break
			}
		}

		chunk := strings.TrimSpace(string(runes[:splitAt]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		runes = runes[splitAt:]
		for len(runes) > 0 && unicodeIsSpace(runes[0]) {
			runes = runes[1:]
		}
	}

	return chunks
}

func unicodeIsSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t'
}
