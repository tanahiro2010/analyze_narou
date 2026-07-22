package main

import (
	"analyze_narou/internal/app"
	"os"
)

func main() {

	config := app.Config{
		os.Getenv("NAROU_URL"),
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	app.Run(config)
}
