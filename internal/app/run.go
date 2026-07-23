package app

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/logger"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	NarouUrl          string
	OpenAIApiKey      string
	DiscordWebhookURL string
}

func Run(config Config, mode narou.RankingMode) {
	var analyzer *analytics.Analyzer
	if config.OpenAIApiKey != "" {
		openaiClient := gpt.NewOpenAIClient(gpt.OpenAIConfig{
			Model:  openai.GPT3Dot5Turbo,
			ApiKey: config.OpenAIApiKey,
		})
		analyzer = analytics.NewAnalyzer(openaiClient)
	} else {
		analyzer = analytics.NewAnalyzer(nil)
	}

	narouClient := narou.NewNarouClient(narou.NarouConfig{
		NarouURL:  config.NarouUrl,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	})

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: config.DiscordWebhookURL,
		Timeout:    10 * time.Second,
	})

	log := logger.NewWebhookLogger(*discordClient)

	var genreAnalyzeResult []analytics.GenreAnalyzeResult

	for _, genre := range narou.BigGenres {
		fmt.Printf("Getting ranking for genre %s with mode %s\n", genre, mode)

		ranking, err := narouClient.GetRankingWithNovelAPI(genre, mode)
		if err != nil {
			fmt.Printf("Error getting ranking: %s\n", err)
			fmt.Println("Continuing to next genre...")
			continue
		}

		fmt.Printf("Ranking for genre %s: %+v\n", genre, *ranking)

		result, err := analyzer.GenreAnalyze(*ranking)
		if err != nil {
			fmt.Printf("Error analyzing genre %s: %s\n", genre, err)
			continue
		}

		genreAnalyzeResult = append(genreAnalyzeResult, result)
		if err := log.GenreAnalyzeResult(result); err != nil {
			fmt.Printf("Error logging genre analysis result: %s\n", err)
		}
	}

	allAnalyzeResult, err := analyzer.AllAnalyze(genreAnalyzeResult)
	if err != nil {
		fmt.Printf("Error analyzing all genres: %s\n", err)
		return
	}

	fmt.Printf("All genres analyzed: %s\n", allAnalyzeResult)
	if err := log.AllAnalyzeResult(allAnalyzeResult); err != nil {
		fmt.Printf("Error logging all analysis result: %s\n", err)
	}
}
