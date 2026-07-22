package app

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/logger"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	NarouUrl          string
	OpenAIApiKey      string
	DiscordWebhookURL string
}

func Run(config Config, mode narou.RankingMode) {
	openaiClient := gpt.NewOpenAIClient(gpt.OpenAIConfig{
		Model:  openai.GPT3Dot5Turbo,
		ApiKey: config.OpenAIApiKey,
	})

	narouClient := narou.NewNarouClient(narou.NarouConfig{
		NarouURL: config.NarouUrl,
	})

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: config.DiscordWebhookURL,
		Timeout:    10,
	})

	analyzer := analytics.NewAnalyzer(*openaiClient)
	log := logger.NewWebhookLogger(*discordClient)

	var genreAnalyzeResult []analytics.GenreAnalyzeResult

	for _, genre := range narou.BigGenres {
		ranking, err := narouClient.GetRanking(genre, mode)
		if err != nil {
			fmt.Printf("Error getting ranking: %s\n", err)
			fmt.Println("Continuing to next genre...")
			continue
		}

		var novels []narou.Novel

		for _, item := range *ranking {
			novel, _ := narouClient.GetNovel(item.Ncode)
			if novel == nil {
				fmt.Printf("Error getting novel for ncode %s\n", item.Ncode)
				continue
			}

			novels = append(novels, *novel)
		}

		result, err := analyzer.GenreAnalyze(novels)
		if err != nil {
			fmt.Printf("Error analyzing genre %s: %s\n", genre, err)
			continue
		}

		genreAnalyzeResult = append(genreAnalyzeResult, result)
		log.GenreAnalyzeResult(result)
	}

	allAnalyzeResult, err := analyzer.AllAnalyze(genreAnalyzeResult)
	if err != nil {
		fmt.Printf("Error analyzing all genres: %s\n", err)
		return
	}

	fmt.Printf("All genres analyzed: %s\n", allAnalyzeResult)
	log.AllAnalyzeResult(allAnalyzeResult)
}
