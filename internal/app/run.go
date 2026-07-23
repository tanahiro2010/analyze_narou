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
		Timeout:    10,
	})

	log := logger.NewWebhookLogger(*discordClient)

	var genreAnalyzeResult []analytics.GenreAnalyzeResult

	for _, genre := range narou.BigGenres {
		date := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
		formatedDate := fmt.Sprintf("%02d", date.Year()) + fmt.Sprintf("%02d", int(date.Month())) + fmt.Sprintf("%02d", date.Day())

		ranking, err := narouClient.GetRanking(genre, formatedDate, mode)
		if err != nil {
			fmt.Printf("Error getting ranking: %s\n", err)
			fmt.Println("Continuing to next genre...")
			continue
		}

		fmt.Printf("Ranking for genre %s: %+v\n", genre, *ranking)

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
