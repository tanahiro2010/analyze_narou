package narou

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
