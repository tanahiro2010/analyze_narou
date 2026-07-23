package main

import (
	"strings"
	"testing"

	"analyze_narou/internal/client/narou"
)

func TestRankingModeFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want narou.RankingMode
	}{
		{name: "default daily", args: nil, want: narou.RankingModeDaily},
		{name: "daily", args: []string{"daily"}, want: narou.RankingModeDaily},
		{name: "daily short", args: []string{"d"}, want: narou.RankingModeDaily},
		{name: "daily uppercase", args: []string{"DAILY"}, want: narou.RankingModeDaily},
		{name: "weekly", args: []string{"weekly"}, want: narou.RankingModeWeekly},
		{name: "weekly short", args: []string{"w"}, want: narou.RankingModeWeekly},
		{name: "quarterly", args: []string{"quarterly"}, want: narou.RankingModeQuarterly},
		{name: "quarter alias", args: []string{"quarter"}, want: narou.RankingModeQuarterly},
		{name: "quarterly short", args: []string{"q"}, want: narou.RankingModeQuarterly},
		{name: "yearly", args: []string{"yearly"}, want: narou.RankingModeYearly},
		{name: "annual alias", args: []string{"annual"}, want: narou.RankingModeYearly},
		{name: "yearly short", args: []string{"y"}, want: narou.RankingModeYearly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rankingModeFromArgs(tt.args)
			if err != nil {
				t.Fatalf("rankingModeFromArgs returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("rankingModeFromArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRankingModeFromArgsReturnsErrorForInvalidMode(t *testing.T) {
	_, err := rankingModeFromArgs([]string{"monthly"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "daily, weekly, quarterly, or yearly") {
		t.Fatalf("error = %q, want usage", err)
	}
}

func TestRankingModeFromArgsReturnsErrorForTooManyArgs(t *testing.T) {
	_, err := rankingModeFromArgs([]string{"daily", "weekly"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "too many arguments") {
		t.Fatalf("error = %q, want too many arguments", err)
	}
}
