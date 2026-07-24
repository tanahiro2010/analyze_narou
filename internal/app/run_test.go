package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"analyze_narou/internal/client/discord"
	"analyze_narou/internal/client/narou"
)

func TestRunFetchesRankingsWithNovelAPIForEachBigGenre(t *testing.T) {
	var novelAPIRequests atomic.Int32
	var activeNovelAPIRequests atomic.Int32
	var maxActiveNovelAPIRequests atomic.Int32
	var webhookRequests atomic.Int32
	var webhookMessagesMu sync.Mutex
	var webhookMessages []discord.WebhookMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/webhook":
			var message discord.WebhookMessage
			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				t.Fatalf("failed to decode webhook payload: %v", err)
			}
			webhookMessagesMu.Lock()
			webhookMessages = append(webhookMessages, message)
			webhookMessagesMu.Unlock()
			webhookRequests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		case "/novelapi/api/":
			novelAPIRequests.Add(1)
			if got := r.URL.Query().Get("out"); got != "json" {
				t.Fatalf("novel out = %q, want json", got)
			}
			if got := r.URL.Query().Get("order"); got != "dailypoint" {
				t.Fatalf("order = %q, want dailypoint", got)
			}
			if got := r.URL.Query().Get("lim"); got != "100" {
				t.Fatalf("lim = %q, want 100", got)
			}

			bigGenre := r.URL.Query().Get("biggenre")
			if bigGenre == "" {
				t.Fatal("biggenre is empty")
			}

			active := activeNovelAPIRequests.Add(1)
			updateMaxAtomicInt32(&maxActiveNovelAPIRequests, active)
			defer activeNovelAPIRequests.Add(-1)
			time.Sleep(20 * time.Millisecond)

			ncode := "N" + bigGenre
			fmt.Fprintf(w, `[
				{"allcount":1},
				{"title":"title %s","ncode":%q,"biggenre":%s,"novel_type":1,"end":1,"daily_point":100}
			]`, ncode, ncode, bigGenre)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	Run(Config{
		NarouUrl:                server.URL + "/",
		NarouUserAgent:          "test",
		NarouRankingLimit:       100,
		GenreAnalyzeConcurrency: 4,
		OpenAIApiKey:            "",
		DiscordWebhookURL:       server.URL + "/webhook",
		DiscordTimeout:          10 * time.Second,
	}, narou.RankingModeDaily)

	if got := novelAPIRequests.Load(); got != int32(len(narou.BigGenres)) {
		t.Fatalf("novel api requests = %d, want %d", got, len(narou.BigGenres))
	}

	if got := maxActiveNovelAPIRequests.Load(); got <= 1 {
		t.Fatalf("max active novel api requests = %d, want concurrent requests", got)
	}

	if got := webhookRequests.Load(); got != int32(len(narou.BigGenres)+2) {
		t.Fatalf("webhook requests = %d, want %d", got, len(narou.BigGenres)+2)
	}

	webhookMessagesMu.Lock()
	firstWebhookContent := webhookMessages[0].Content
	webhookMessagesMu.Unlock()
	if firstWebhookContent != "# デイリーランキング分析" {
		t.Fatalf("first webhook content = %q, want daily h1", firstWebhookContent)
	}
}

func TestGenreLogNameIncludesNameAndCode(t *testing.T) {
	got := genreLogName(narou.BigGenreFantasy)
	want := "ファンタジー(2)"

	if got != want {
		t.Fatalf("genreLogName() = %q, want %q", got, want)
	}
}

func TestRankingModeLogTitle(t *testing.T) {
	tests := []struct {
		mode narou.RankingMode
		want string
	}{
		{mode: narou.RankingModeDaily, want: "デイリーランキング分析"},
		{mode: narou.RankingModeWeekly, want: "ウィークリーランキング分析"},
		{mode: narou.RankingModeQuarterly, want: "四半期ランキング分析"},
		{mode: narou.RankingModeYearly, want: "年間ランキング分析"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := rankingModeLogTitle(tt.mode); got != tt.want {
				t.Fatalf("rankingModeLogTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func updateMaxAtomicInt32(max *atomic.Int32, value int32) {
	for {
		current := max.Load()
		if value <= current {
			return
		}
		if max.CompareAndSwap(current, value) {
			return
		}
	}
}
