package main

import (
	"testing"

	"analyze_narou/internal/client/narou"
)

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("NAROU_URL", "https://api.example.test/")
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.example.test/webhook")

	config := configFromEnv()

	if config.NarouUrl != "https://api.example.test/" {
		t.Fatalf("NarouUrl = %q", config.NarouUrl)
	}

	if config.OpenAIApiKey != "openai-key" {
		t.Fatalf("OpenAIApiKey = %q", config.OpenAIApiKey)
	}

	if config.DiscordWebhookURL != "https://discord.example.test/webhook" {
		t.Fatalf("DiscordWebhookURL = %q", config.DiscordWebhookURL)
	}
}

func TestDefaultRankingMode(t *testing.T) {
	if got := defaultRankingMode(); got != narou.RankingModeDaily {
		t.Fatalf("defaultRankingMode() = %q, want %q", got, narou.RankingModeDaily)
	}
}
