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
	var activeNovelAPIRequests atomic.Int32
	var maxActiveNovelAPIRequests atomic.Int32
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

	if got := webhookRequests.Load(); got != int32(len(narou.BigGenres)+1) {
		t.Fatalf("webhook requests = %d, want %d", got, len(narou.BigGenres)+1)
	}
}

func TestGenreLogNameIncludesNameAndCode(t *testing.T) {
	got := genreLogName(narou.BigGenreFantasy)
	want := "ファンタジー(2)"

	if got != want {
		t.Fatalf("genreLogName() = %q, want %q", got, want)
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
