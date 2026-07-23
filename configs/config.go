package configs

import (
	"analyze_narou/internal/app"
	"os"
)

var CONFIG = Load()

func Load() app.Config {
	return app.Config{
		NarouUrl:          os.Getenv("NAROU_URL"),
		OpenAIApiKey:      os.Getenv("OPENAI_API_KEY"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}
}
