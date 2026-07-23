package configs

import (
	"os"
	"strconv"
	"time"

	"github.com/sashabaranov/go-openai"
)

var CONFIG = Load()

const (
	DefaultNarouURL          = "https://api.syosetu.com/"
	DefaultNarouUserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
	DefaultNarouRankingLimit = 100
	DefaultOpenAIModel       = openai.GPT3Dot5Turbo
	DefaultDiscordTimeout    = 10 * time.Second
)

type Config struct {
	NarouUrl          string
	NarouUserAgent    string
	NarouRankingLimit int
	OpenAIApiKey      string
	OpenAIModel       string
	DiscordWebhookURL string
	DiscordTimeout    time.Duration
}

func Load() Config {
	return Config{
		NarouUrl:          stringFromEnv("NAROU_URL", DefaultNarouURL),
		NarouUserAgent:    stringFromEnv("NAROU_USER_AGENT", DefaultNarouUserAgent),
		NarouRankingLimit: intFromEnv("NAROU_RANKING_LIMIT", DefaultNarouRankingLimit),
		OpenAIApiKey:      os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:       stringFromEnv("OPENAI_MODEL", DefaultOpenAIModel),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		DiscordTimeout:    durationFromEnv("DISCORD_TIMEOUT", DefaultDiscordTimeout),
	}
}

func stringFromEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func intFromEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
