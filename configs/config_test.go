package configs

import "testing"

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("NAROU_URL", "https://api.example.test/")
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.example.test/webhook")

	config := Load()

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
