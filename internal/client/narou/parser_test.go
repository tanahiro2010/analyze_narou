package narou

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNovelUnmarshalJSONAcceptsNovelTypeKeys(t *testing.T) {
	tests := []struct {
		name string
		body string
		want NovelType
	}{
		{name: "normal key", body: `{"title":"normal","novel_type":1}`, want: NovelTypeSerial},
		{name: "of key", body: `{"title":"of","noveltype":2}`, want: NovelTypeShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var novel Novel
			if err := json.Unmarshal([]byte(tt.body), &novel); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}

			if novel.NovelType != tt.want {
				t.Fatalf("NovelType = %v, want %v", novel.NovelType, tt.want)
			}
		})
	}
}

func TestResponseUnmarshalJSONDecodesMetadataAndSkipsNullNovels(t *testing.T) {
	body := `[
		{"allcount":2},
		null,
		{"title":"first","ncode":"N1","novel_type":1},
		{"title":"second","ncode":"N2","noveltype":2}
	]`

	var response Response
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if response.AllCount != 2 {
		t.Fatalf("AllCount = %d, want 2", response.AllCount)
	}

	if len(response.Novels) != 2 {
		t.Fatalf("len(Novels) = %d, want 2", len(response.Novels))
	}

	if response.Novels[0].NCode != "N1" || response.Novels[1].NCode != "N2" {
		t.Fatalf("unexpected novels: %+v", response.Novels)
	}
}

func TestResponseUnmarshalJSONHandlesEmptyArray(t *testing.T) {
	var response Response
	if err := json.Unmarshal([]byte(`[]`), &response); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if response.AllCount != 0 {
		t.Fatalf("AllCount = %d, want 0", response.AllCount)
	}

	if response.Novels != nil {
		t.Fatalf("Novels = %#v, want nil", response.Novels)
	}
}

func TestResponseUnmarshalJSONReturnsIndexedNovelError(t *testing.T) {
	var response Response
	err := json.Unmarshal([]byte(`[{"allcount":1}, []]`), &response)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "response index 1") {
		t.Fatalf("error = %q, want response index", err)
	}
}

func TestParseTime(t *testing.T) {
	got, err := ParseTime("2026-07-23 12:34:56")
	if err != nil {
		t.Fatalf("ParseTime returned error: %v", err)
	}

	if got.Location().String() != "Asia/Tokyo" {
		t.Fatalf("location = %s, want Asia/Tokyo", got.Location())
	}

	if got.Year() != 2026 || got.Month() != 7 || got.Day() != 23 ||
		got.Hour() != 12 || got.Minute() != 34 || got.Second() != 56 {
		t.Fatalf("unexpected parsed time: %s", got)
	}
}

func TestParseTimeReturnsErrorForInvalidInput(t *testing.T) {
	if _, err := ParseTime("not-a-time"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRankingResultUnmarshalJSON(t *testing.T) {
	var ranking RankingResult
	if err := json.Unmarshal([]byte(`[{"ncode":"N2","pt":20,"rank":2},{"ncode":"N1","pt":30,"rank":1}]`), &ranking); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if len(ranking) != 2 {
		t.Fatalf("len(ranking) = %d, want 2", len(ranking))
	}

	if ranking[0].Ncode != "N2" || ranking[0].Pt != 20 || ranking[0].Rank != 2 {
		t.Fatalf("unexpected first ranking item: %+v", ranking[0])
	}
}
