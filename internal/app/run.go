package app

import (
	"analyze_narou/configs"
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"analyze_narou/internal/logger"
	"fmt"
	"sync"
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

	concurrency := config.GenreAnalyzeConcurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	genreAnalyzeResults := make(chan analytics.GenreAnalyzeResult, len(narou.BigGenres))
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, genre := range narou.BigGenres {
		genre := genre
		wg.Add(1)

		go func() {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			result, ok := analyzeGenre(genre, mode, narouClient, analyzer, log)
			if ok {
				genreAnalyzeResults <- result
			}
		}()
	}

	wg.Wait()
	close(genreAnalyzeResults)

	var genreAnalyzeResult []analytics.GenreAnalyzeResult
	for result := range genreAnalyzeResults {
		genreAnalyzeResult = append(genreAnalyzeResult, result)
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

func analyzeGenre(
	genre narou.BigGenre,
	mode narou.RankingMode,
	narouClient *narou.NarouClient,
	analyzer *analytics.Analyzer,
	log *logger.WebhookLogger,
) (analytics.GenreAnalyzeResult, bool) {
	genreName := genreLogName(genre)
	fmt.Printf("Getting ranking for genre %s with mode %s\n", genreName, mode)

	ranking, err := narouClient.GetRankingWithNovelAPI(genre, mode)
	if err != nil {
		fmt.Printf("Error getting ranking for genre %s: %s\n", genreName, err)
		fmt.Println("Continuing to next genre...")
		return analytics.GenreAnalyzeResult{}, false
	}

	fmt.Printf("Ranking for genre %s: %+v\n", genreName, *ranking)

	result, err := analyzer.GenreAnalyze(*ranking, genre.String())
	if err != nil {
		fmt.Printf("Error analyzing genre %s: %s\n", genreName, err)
		return analytics.GenreAnalyzeResult{}, false
	}

	if err := log.GenreAnalyzeResult(result); err != nil {
		fmt.Printf("Error logging genre analysis result for genre %s: %s\n", genreName, err)
	}

	return result, true
}

func genreLogName(genre narou.BigGenre) string {
	return fmt.Sprintf("%s(%d)", genre.String(), genre)
}
