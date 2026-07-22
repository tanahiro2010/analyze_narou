package logger

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
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

	return nil
}

func (w *WebhookLogger) GenreAnalyzeResult(ctx analytics.GenreAnalyzeResult) error {
	message := ""

	return w.Log(message)
}

func (w *WebhookLogger) AllAnalyzeResult(ctx analytics.AllAnalyzeResult) error {
	message := ""

	return w.Log(message)
}
