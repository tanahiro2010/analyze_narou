package narou

import (
	"sort"
	"testing"
)

func TestBinaryFlagBool(t *testing.T) {
	tests := []struct {
		name string
		flag BinaryFlag
		want bool
	}{
		{name: "enabled", flag: FlagEnabled, want: true},
		{name: "disabled", flag: FlagDisabled, want: false},
		{name: "unknown", flag: BinaryFlag(2), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flag.Bool(); got != tt.want {
				t.Fatalf("Bool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigGenreStringAndIsNone(t *testing.T) {
	tests := []struct {
		genre  BigGenre
		label  string
		isNone bool
	}{
		{genre: BigGenreNone, label: "未選択", isNone: true},
		{genre: BigGenreRomance, label: "恋愛"},
		{genre: BigGenreFantasy, label: "ファンタジー"},
		{genre: BigGenreLiterature, label: "文芸"},
		{genre: BigGenreSF, label: "SF"},
		{genre: BigGenreNonGenre, label: "ノンジャンル"},
		{genre: BigGenreOther, label: "その他"},
		{genre: BigGenre(999), label: "不明"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			if got := tt.genre.String(); got != tt.label {
				t.Fatalf("String() = %q, want %q", got, tt.label)
			}

			if got := tt.genre.IsNone(); got != tt.isNone {
				t.Fatalf("IsNone() = %v, want %v", got, tt.isNone)
			}
		})
	}
}

func TestGenreStringAndIsNone(t *testing.T) {
	tests := []struct {
		genre  Genre
		label  string
		isNone bool
	}{
		{genre: GenreNone, label: "未選択", isNone: true},
		{genre: GenreIsekaiRomance, label: "異世界〔恋愛〕"},
		{genre: GenreRealRomance, label: "現実世界〔恋愛〕"},
		{genre: GenreHighFantasy, label: "ハイファンタジー〔ファンタジー〕"},
		{genre: GenreLowFantasy, label: "ローファンタジー〔ファンタジー〕"},
		{genre: GenrePureLiterature, label: "純文学〔文芸〕"},
		{genre: GenreHumanDrama, label: "ヒューマンドラマ〔文芸〕"},
		{genre: GenreHistory, label: "歴史〔文芸〕"},
		{genre: GenreMystery, label: "推理〔文芸〕"},
		{genre: GenreHorror, label: "ホラー〔文芸〕"},
		{genre: GenreAction, label: "アクション〔文芸〕"},
		{genre: GenreComedy, label: "コメディー〔文芸〕"},
		{genre: GenreVRGame, label: "VRゲーム〔SF〕"},
		{genre: GenreSpace, label: "宇宙〔SF〕"},
		{genre: GenreScienceFiction, label: "空想科学〔SF〕"},
		{genre: GenrePanic, label: "パニック〔SF〕"},
		{genre: GenreNonGenre, label: "ノンジャンル〔ノンジャンル〕"},
		{genre: GenreFairyTale, label: "童話〔その他〕"},
		{genre: GenrePoetry, label: "詩〔その他〕"},
		{genre: GenreEssay, label: "エッセイ〔その他〕"},
		{genre: GenreReplay, label: "リプレイ〔その他〕"},
		{genre: GenreOther, label: "その他〔その他〕"},
		{genre: Genre(999999), label: "不明"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			if got := tt.genre.String(); got != tt.label {
				t.Fatalf("String() = %q, want %q", got, tt.label)
			}

			if got := tt.genre.IsNone(); got != tt.isNone {
				t.Fatalf("IsNone() = %v, want %v", got, tt.isNone)
			}
		})
	}
}

func TestNovelTypeString(t *testing.T) {
	tests := []struct {
		novelType NovelType
		want      string
	}{
		{novelType: NovelTypeSerial, want: "連載"},
		{novelType: NovelTypeShort, want: "短編"},
		{novelType: NovelType(999), want: "不明"},
	}

	for _, tt := range tests {
		if got := tt.novelType.String(); got != tt.want {
			t.Fatalf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestEndStatusString(t *testing.T) {
	tests := []struct {
		status EndStatus
		want   string
	}{
		{status: EndCompleted, want: "完結済み"},
		{status: EndOngoing, want: "連載中"},
		{status: EndStatus(999), want: "不明"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Fatalf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestNovelStateHelpers(t *testing.T) {
	tests := []struct {
		name        string
		novel       Novel
		isShort     bool
		isSerial    bool
		isCompleted bool
		isOngoing   bool
	}{
		{
			name:        "short",
			novel:       Novel{NovelType: NovelTypeShort, End: EndCompleted},
			isShort:     true,
			isCompleted: false,
		},
		{
			name:        "completed serial",
			novel:       Novel{NovelType: NovelTypeSerial, End: EndCompleted},
			isSerial:    true,
			isCompleted: true,
		},
		{
			name:      "ongoing serial",
			novel:     Novel{NovelType: NovelTypeSerial, End: EndOngoing},
			isSerial:  true,
			isOngoing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.novel.IsShort(); got != tt.isShort {
				t.Fatalf("IsShort() = %v, want %v", got, tt.isShort)
			}
			if got := tt.novel.IsSerial(); got != tt.isSerial {
				t.Fatalf("IsSerial() = %v, want %v", got, tt.isSerial)
			}
			if got := tt.novel.IsCompletedSerial(); got != tt.isCompleted {
				t.Fatalf("IsCompletedSerial() = %v, want %v", got, tt.isCompleted)
			}
			if got := tt.novel.IsOngoingSerial(); got != tt.isOngoing {
				t.Fatalf("IsOngoingSerial() = %v, want %v", got, tt.isOngoing)
			}
		})
	}
}

func TestRankingResultSortsByRank(t *testing.T) {
	ranking := RankingResult{
		{Ncode: "N3", Rank: 3},
		{Ncode: "N1", Rank: 1},
		{Ncode: "N2", Rank: 2},
	}

	sort.Sort(ranking)

	got := []string{ranking[0].Ncode, ranking[1].Ncode, ranking[2].Ncode}
	want := []string{"N1", "N2", "N3"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted ncode[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
