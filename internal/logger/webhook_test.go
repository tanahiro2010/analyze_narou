package logger

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"testing"
)

func TestNewWebhookLogger(t *testing.T) {
	discordClient := discord.DiscordClient{}

	logger := NewWebhookLogger(discordClient)
	if logger == nil {
		t.Fatal("logger is nil")
	}
}

func TestWebhookLoggerMethodsReturnNil(t *testing.T) {
	logger := NewWebhookLogger(discord.DiscordClient{})

	if err := logger.Log("message"); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}

	if err := logger.GenreAnalyzeResult(analytics.GenreAnalyzeResult{}); err != nil {
		t.Fatalf("GenreAnalyzeResult returned error: %v", err)
	}

	if err := logger.AllAnalyzeResult(analytics.AllAnalyzeResult{}); err != nil {
		t.Fatalf("AllAnalyzeResult returned error: %v", err)
	}
}
