package app

import (
	"analyze_narou/internal/client/gpt"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	NarouUrl          string
	OpenAIApiKey      string
	DiscordWebhookURL string
}

func Run(config Config) error {
	openaiClient := gpt.NewOpenAIClient(gpt.OpenAIConfig{
		Model:  openai.GPT3Dot5Turbo,
		ApiKey: config.OpenAIApiKey,
	})

	narouClient := narou.NewNarouClient(narou.NarouConfig{
		Endpoint: config.NarouUrl,
	})

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: config.DiscordWebhookURL,
	})
	return nil
}
