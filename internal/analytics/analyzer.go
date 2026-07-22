package analytics

import (
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
)

type GenreAnalyzeResult struct {
}
type AllAnalyzeResult struct {
}

type Analyzer struct {
	OpenAIClient gpt.OpenAIClient
}

func NewAnalyzer(openAIClient gpt.OpenAIClient) *Analyzer {
	return &Analyzer{
		OpenAIClient: openAIClient,
	}
}

func (a *Analyzer) GenreAnalyze(ctx []narou.Novel) (GenreAnalyzeResult, error) {

	return GenreAnalyzeResult{}, nil
}

func (a *Analyzer) AllAnalyze(ctx []GenreAnalyzeResult) (AllAnalyzeResult, error) {

	return AllAnalyzeResult{}, nil
}
