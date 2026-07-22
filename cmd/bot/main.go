package main

import (
	"analyze_narou/internal/app"
	"analyze_narou/internal/client/narou"
	"os"
)

func main() {
	config := app.Config{
		NarouUrl:          os.Getenv("NAROU_URL"),
		OpenAIApiKey:      os.Getenv("OPENAI_API_KEY"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	mode := narou.RankingModeDaily // You can change this to the desired ranking mode

	app.Run(config, mode)
}
