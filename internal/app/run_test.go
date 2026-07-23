package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"analyze_narou/internal/client/narou"
)

func TestRunFetchesRankingsWithNovelAPIForEachBigGenre(t *testing.T) {
	var novelAPIRequests atomic.Int32
	var webhookRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/webhook":
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
		NarouUrl:          server.URL + "/",
		NarouUserAgent:    "test",
		NarouRankingLimit: 100,
		OpenAIApiKey:      "",
		DiscordWebhookURL: server.URL + "/webhook",
		DiscordTimeout:    10 * time.Second,
	}, narou.RankingModeDaily)

	if got := novelAPIRequests.Load(); got != int32(len(narou.BigGenres)) {
		t.Fatalf("novel api requests = %d, want %d", got, len(narou.BigGenres))
	}

	if got := webhookRequests.Load(); got != int32(len(narou.BigGenres)+1) {
		t.Fatalf("webhook requests = %d, want %d", got, len(narou.BigGenres)+1)
	}
}
