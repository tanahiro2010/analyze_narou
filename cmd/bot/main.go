package main

import (
	"analyze_narou/internal/app"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/utils"
	"os"
)

func main() {
	utils.LoadDotEnv(".env")

	config := app.Config{
		NarouUrl:          os.Getenv("NAROU_URL"),
		OpenAIApiKey:      os.Getenv("OPENAI_API_KEY"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	mode := narou.RankingModeDaily // cmdのargから指定できるように

	app.Run(config, mode)
}
