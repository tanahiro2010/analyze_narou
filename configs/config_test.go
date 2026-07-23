package configs

import (
	"testing"
	"time"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("NAROU_URL", "https://api.example.test/")
	t.Setenv("NAROU_USER_AGENT", "test-agent")
	t.Setenv("NAROU_RANKING_LIMIT", "50")
	t.Setenv("OPENAI_API_KEY", "openai-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.example.test/webhook")
	t.Setenv("DISCORD_TIMEOUT", "3s")

	config := Load()

	if config.NarouUrl != "https://api.example.test/" {
		t.Fatalf("NarouUrl = %q", config.NarouUrl)
	}

	if config.NarouUserAgent != "test-agent" {
		t.Fatalf("NarouUserAgent = %q", config.NarouUserAgent)
	}

	if config.NarouRankingLimit != 50 {
		t.Fatalf("NarouRankingLimit = %d", config.NarouRankingLimit)
	}

	if config.OpenAIApiKey != "openai-key" {
		t.Fatalf("OpenAIApiKey = %q", config.OpenAIApiKey)
	}

	if config.OpenAIModel != "test-model" {
		t.Fatalf("OpenAIModel = %q", config.OpenAIModel)
	}

	if config.DiscordWebhookURL != "https://discord.example.test/webhook" {
		t.Fatalf("DiscordWebhookURL = %q", config.DiscordWebhookURL)
	}

	if config.DiscordTimeout != 3*time.Second {
		t.Fatalf("DiscordTimeout = %s", config.DiscordTimeout)
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("NAROU_URL", "")
	t.Setenv("NAROU_USER_AGENT", "")
	t.Setenv("NAROU_RANKING_LIMIT", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("DISCORD_TIMEOUT", "")

	config := Load()

	if config.NarouUrl != DefaultNarouURL {
		t.Fatalf("NarouUrl = %q, want %q", config.NarouUrl, DefaultNarouURL)
	}

	if config.NarouUserAgent != DefaultNarouUserAgent {
		t.Fatalf("NarouUserAgent = %q, want %q", config.NarouUserAgent, DefaultNarouUserAgent)
	}

	if config.NarouRankingLimit != DefaultNarouRankingLimit {
		t.Fatalf("NarouRankingLimit = %d, want %d", config.NarouRankingLimit, DefaultNarouRankingLimit)
	}

	if config.OpenAIModel != DefaultOpenAIModel {
		t.Fatalf("OpenAIModel = %q, want %q", config.OpenAIModel, DefaultOpenAIModel)
	}

	if config.DiscordTimeout != DefaultDiscordTimeout {
		t.Fatalf("DiscordTimeout = %s, want %s", config.DiscordTimeout, DefaultDiscordTimeout)
	}
}

func TestLoadUsesDefaultsForInvalidValues(t *testing.T) {
	t.Setenv("NAROU_RANKING_LIMIT", "not-an-int")
	t.Setenv("DISCORD_TIMEOUT", "not-a-duration")

	config := Load()

	if config.NarouRankingLimit != DefaultNarouRankingLimit {
		t.Fatalf("NarouRankingLimit = %d, want %d", config.NarouRankingLimit, DefaultNarouRankingLimit)
	}

	if config.DiscordTimeout != DefaultDiscordTimeout {
		t.Fatalf("DiscordTimeout = %s, want %s", config.DiscordTimeout, DefaultDiscordTimeout)
	}
}
