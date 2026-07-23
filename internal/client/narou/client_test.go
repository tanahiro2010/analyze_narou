package narou

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetNovelDecodesNarouArrayResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/novelapi/api/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("ncode"); got != "N3973LX" {
			t.Fatalf("unexpected ncode: %s", got)
		}

		fmt.Fprint(w, `[
			{"allcount":1},
			{
				"title":"Test Novel",
				"ncode":"N3973LX",
				"userid":123,
				"writer":"tester",
				"story":"story",
				"keyword":"keyword",
				"biggenre":2,
				"genre":201,
				"novel_type":1,
				"end":1,
				"isstop":0,
				"general_all_no":10,
				"length":1000,
				"time":5,
				"isr15":0,
				"isbl":0,
				"isgl":0,
				"iszankoku":0,
				"istensei":0,
				"istenni":0,
				"global_point":100,
				"daily_point":10,
				"weekly_point":20,
				"monthly_point":30,
				"quarter_point":40,
				"yearly_point":50,
				"fav_novel_cnt":3,
				"impression_cnt":4,
				"review_cnt":5,
				"all_point":6,
				"all_hyoka_cnt":7,
				"sasie_cnt":8,
				"kaiwaritu":9,
				"novelupdated_at":"2026-07-23 12:00:00",
				"updated_at":"2026-07-23 12:00:00"
			}
		]`)
	}))
	defer server.Close()

	client := NewNarouClient(NarouConfig{
		NarouURL:  server.URL + "/",
		UserAgent: "test",
	})

	novel, err := client.GetNovel("N3973LX")
	if err != nil {
		t.Fatalf("GetNovel returned error: %v", err)
	}

	if novel.NCode != "N3973LX" {
		t.Fatalf("unexpected ncode: %s", novel.NCode)
	}

	if novel.Title != "Test Novel" {
		t.Fatalf("unexpected title: %s", novel.Title)
	}

	if novel.NovelType != NovelTypeSerial {
		t.Fatalf("unexpected novel type: %v", novel.NovelType)
	}
}

func TestGetNovelReturnsErrorForEmptyNovelResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"allcount":0}]`)
	}))
	defer server.Close()

	client := NewNarouClient(NarouConfig{
		NarouURL:  server.URL + "/",
		UserAgent: "test",
	})

	if _, err := client.GetNovel("N0000AA"); err == nil {
		t.Fatal("expected error for empty novel response")
	}
}

func TestGetNovelReturnsErrorForInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{`)
	}))
	defer server.Close()

	client := NewNarouClient(NarouConfig{
		NarouURL:  server.URL + "/",
		UserAgent: "test",
	})

	if _, err := client.GetNovel("N0000AA"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetRankingDecodesRankingResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rank/rankget/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("biggenre"); got != "2" {
			t.Fatalf("biggenre = %q, want 2", got)
		}

		if got := r.URL.Query().Get("rtype"); got != "20260723-d" {
			t.Fatalf("rtype = %q, want 20260723-d", got)
		}

		fmt.Fprint(w, `[{"ncode":"N1","pt":100,"rank":1},{"ncode":"N2","pt":90,"rank":2}]`)
	}))
	defer server.Close()

	client := NewNarouClient(NarouConfig{
		NarouURL:  server.URL + "/",
		UserAgent: "test",
	})

	ranking, err := client.GetRanking(BigGenreFantasy, "20260723", RankingModeDaily)
	if err != nil {
		t.Fatalf("GetRanking returned error: %v", err)
	}

	if len(*ranking) != 2 {
		t.Fatalf("len(ranking) = %d, want 2", len(*ranking))
	}

	if (*ranking)[0].Ncode != "N1" || (*ranking)[0].Pt != 100 || (*ranking)[0].Rank != 1 {
		t.Fatalf("unexpected first ranking: %+v", (*ranking)[0])
	}
}

func TestGetRankingReturnsErrorForInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{`)
	}))
	defer server.Close()

	client := NewNarouClient(NarouConfig{
		NarouURL:  server.URL + "/",
		UserAgent: "test",
	})

	_, err := client.GetRanking(BigGenreFantasy, "20260723", RankingModeDaily)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "unexpected end of JSON input") {
		t.Fatalf("error = %q, want unexpected end of JSON input", err)
	}
}
