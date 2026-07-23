package app

import (
	"analyze_narou/configs"
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/logger"
	"fmt"
)

type Config = configs.Config

func Run(config Config, mode narou.RankingMode) {
	var analyzer *analytics.Analyzer
	if config.OpenAIApiKey != "" {
		openaiClient := gpt.NewOpenAIClient(gpt.OpenAIConfig{
			ApiKey:  config.OpenAIApiKey,
			BaseURL: config.OpenAIBaseURL,
			Model:   config.OpenAIModel,
		})
		analyzer = analytics.NewAnalyzer(openaiClient)
	} else {
		analyzer = analytics.NewAnalyzer(nil)
	}

	narouClient := narou.NewNarouClient(narou.NarouConfig{
		NarouURL:     config.NarouUrl,
		UserAgent:    config.NarouUserAgent,
		RankingLimit: config.NarouRankingLimit,
	})

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: config.DiscordWebhookURL,
		Timeout:    config.DiscordTimeout,
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
