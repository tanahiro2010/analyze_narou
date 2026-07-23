package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"analyze_narou/internal/client/narou"
)

func TestRunFetchesRankingsAndNovelsForEachBigGenre(t *testing.T) {
	var rankingRequests atomic.Int32
	var novelRequests atomic.Int32
	var webhookRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/webhook":
			webhookRequests.Add(1)
			w.WriteHeader(http.StatusNoContent)
		case "/rank/rankget/":
			rankingRequests.Add(1)
			if got := r.URL.Query().Get("out"); got != "json" {
				t.Fatalf("ranking out = %q, want json", got)
			}
			wantRType := time.Now().Format("20060102") + "-d"
			if got := r.URL.Query().Get("rtype"); got != wantRType {
				t.Fatalf("rtype = %q, want %s", got, wantRType)
			}

			bigGenre := r.URL.Query().Get("biggenre")
			ncode := "N" + bigGenre
			fmt.Fprintf(w, `[{"ncode":%q,"pt":100,"rank":1}]`, ncode)
		case "/novelapi/api/":
			novelRequests.Add(1)
			if got := r.URL.Query().Get("out"); got != "json" {
				t.Fatalf("novel out = %q, want json", got)
			}

			ncode := r.URL.Query().Get("ncode")
			if _, err := strconv.Atoi(ncode[1:]); err != nil {
				t.Fatalf("unexpected ncode: %q", ncode)
			}

			fmt.Fprintf(w, `[
				{"allcount":1},
				{"title":"title %s","ncode":%q,"novel_type":1,"end":1}
			]`, ncode, ncode)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	Run(Config{
		NarouUrl:          server.URL + "/",
		OpenAIApiKey:      "",
		DiscordWebhookURL: server.URL + "/webhook",
	}, narou.RankingModeDaily)

	if got := rankingRequests.Load(); got != int32(len(narou.BigGenres)) {
		t.Fatalf("ranking requests = %d, want %d", got, len(narou.BigGenres))
	}

	if got := novelRequests.Load(); got != int32(len(narou.BigGenres)) {
		t.Fatalf("novel requests = %d, want %d", got, len(narou.BigGenres))
	}

	if got := webhookRequests.Load(); got != int32(len(narou.BigGenres)+1) {
		t.Fatalf("webhook requests = %d, want %d", got, len(narou.BigGenres)+1)
	}
}
