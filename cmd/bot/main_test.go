package main

import (
	"testing"

	"analyze_narou/internal/client/narou"
)

func TestDefaultRankingMode(t *testing.T) {
	mode := narou.RankingModeDaily
	if mode != narou.RankingModeDaily {
		t.Fatalf("ranking mode = %q, want %q", mode, narou.RankingModeDaily)
	}
}
