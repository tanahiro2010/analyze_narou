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
		resp, err := w.DiscordClient.SendMessage(discord.WebhookMessage{
			Username: "Narou Analyzer",
			Content:  content,
		})
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
	}
	return nil
}

func (w *WebhookLogger) GenreAnalyzeResult(ctx analytics.GenreAnalyzeResult) error {
	message := "## ジャンル別ランキング分析\n" + ctx.String()

	return w.Log(message)
}

func (w *WebhookLogger) AllAnalyzeResult(ctx analytics.AllAnalyzeResult) error {
	message := "## 全体ランキング分析\n" + ctx.String()

	return w.Log(message)
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
