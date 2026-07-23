package analytics

import (
	"analyze_narou/internal/client/gpt"
	"analyze_narou/internal/client/narou"
	"testing"
)

func TestNewAnalyzer(t *testing.T) {
	openAIClient := gpt.OpenAIClient{}

	analyzer := NewAnalyzer(openAIClient)
	if analyzer == nil {
		t.Fatal("analyzer is nil")
	}
}

func TestGenreAnalyzeReturnsEmptyResult(t *testing.T) {
	analyzer := NewAnalyzer(gpt.OpenAIClient{})

	result, err := analyzer.GenreAnalyze([]narou.Novel{{NCode: "N1"}})
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	if result != (GenreAnalyzeResult{}) {
		t.Fatalf("result = %+v, want empty result", result)
	}
}

func TestAllAnalyzeReturnsEmptyResult(t *testing.T) {
	analyzer := NewAnalyzer(gpt.OpenAIClient{})

	result, err := analyzer.AllAnalyze([]GenreAnalyzeResult{{}})
	if err != nil {
		t.Fatalf("AllAnalyze returned error: %v", err)
	}

	if result != (AllAnalyzeResult{}) {
		t.Fatalf("result = %+v, want empty result", result)
	}
}
