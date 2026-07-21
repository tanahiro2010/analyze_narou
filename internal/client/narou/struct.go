package narou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const narouTimeLayout = "2006-01-02 15:04:05"

// BinaryFlag は、なろうAPIで使用される0または1のフラグです。
type BinaryFlag int

const (
	FlagDisabled BinaryFlag = 0
	FlagEnabled  BinaryFlag = 1
)

// Bool は、フラグが有効ならtrueを返します。
func (f BinaryFlag) Bool() bool {
	return f == FlagEnabled
}

// BigGenre は、なろうAPIの大ジャンルコードです。
type BigGenre int

const (
	BigGenreNone       BigGenre = 0
	BigGenreRomance    BigGenre = 1
	BigGenreFantasy    BigGenre = 2
	BigGenreLiterature BigGenre = 3
	BigGenreSF         BigGenre = 4
	BigGenreNonGenre   BigGenre = 98
	BigGenreOther      BigGenre = 99
)

// String は、大ジャンルの日本語名を返します。
func (g BigGenre) String() string {
	switch g {
	case BigGenreNone:
		return "未選択"
	case BigGenreRomance:
		return "恋愛"
	case BigGenreFantasy:
		return "ファンタジー"
	case BigGenreLiterature:
		return "文芸"
	case BigGenreSF:
		return "SF"
	case BigGenreNonGenre:
		return "ノンジャンル"
	case BigGenreOther:
		return "その他"
	default:
		return "不明"
	}
}

// Genre は、なろうAPIのジャンルコードです。
type Genre int

const (
	GenreNone Genre = 0

	GenreIsekaiRomance Genre = 101
	GenreRealRomance   Genre = 102

	GenreHighFantasy Genre = 201
	GenreLowFantasy  Genre = 202

	GenrePureLiterature Genre = 301
	GenreHumanDrama     Genre = 302
	GenreHistory        Genre = 303
	GenreMystery        Genre = 304
	GenreHorror         Genre = 305
	GenreAction         Genre = 306
	GenreComedy         Genre = 307

	GenreVRGame         Genre = 401
	GenreSpace          Genre = 402
	GenreScienceFiction Genre = 403
	GenrePanic          Genre = 404

	GenreNonGenre Genre = 9801

	GenreFairyTale Genre = 9901
	GenrePoetry    Genre = 9902
	GenreEssay     Genre = 9903
	GenreReplay    Genre = 9904
	GenreOther     Genre = 9999
)

// String は、ジャンルの日本語名を返します。
func (g Genre) String() string {
	switch g {
	case GenreNone:
		return "未選択"
	case GenreIsekaiRomance:
		return "異世界〔恋愛〕"
	case GenreRealRomance:
		return "現実世界〔恋愛〕"
	case GenreHighFantasy:
		return "ハイファンタジー〔ファンタジー〕"
	case GenreLowFantasy:
		return "ローファンタジー〔ファンタジー〕"
	case GenrePureLiterature:
		return "純文学〔文芸〕"
	case GenreHumanDrama:
		return "ヒューマンドラマ〔文芸〕"
	case GenreHistory:
		return "歴史〔文芸〕"
	case GenreMystery:
		return "推理〔文芸〕"
	case GenreHorror:
		return "ホラー〔文芸〕"
	case GenreAction:
		return "アクション〔文芸〕"
	case GenreComedy:
		return "コメディー〔文芸〕"
	case GenreVRGame:
		return "VRゲーム〔SF〕"
	case GenreSpace:
		return "宇宙〔SF〕"
	case GenreScienceFiction:
		return "空想科学〔SF〕"
	case GenrePanic:
		return "パニック〔SF〕"
	case GenreNonGenre:
		return "ノンジャンル〔ノンジャンル〕"
	case GenreFairyTale:
		return "童話〔その他〕"
	case GenrePoetry:
		return "詩〔その他〕"
	case GenreEssay:
		return "エッセイ〔その他〕"
	case GenreReplay:
		return "リプレイ〔その他〕"
	case GenreOther:
		return "その他〔その他〕"
	default:
		return "不明"
	}
}

// NovelType は、作品が連載か短編かを表します。
type NovelType int

const (
	NovelTypeSerial NovelType = 1
	NovelTypeShort  NovelType = 2
)

// String は、作品タイプの日本語名を返します。
func (t NovelType) String() string {
	switch t {
	case NovelTypeSerial:
		return "連載"
	case NovelTypeShort:
		return "短編"
	default:
		return "不明"
	}
}

// EndStatus は、連載作品の完結状態を表します。
//
// 短編作品もEndCompletedとして返されるため、短編か完結済み連載かを
// 判定する場合はNovelTypeと組み合わせて使用します。
type EndStatus int

const (
	EndCompleted EndStatus = 0
	EndOngoing   EndStatus = 1
)

// String は、完結状態の日本語名を返します。
func (s EndStatus) String() string {
	switch s {
	case EndCompleted:
		return "完結済み"
	case EndOngoing:
		return "連載中"
	default:
		return "不明"
	}
}

// Response は、なろう小説APIのレスポンス全体を表します。
//
// なろう小説APIのJSONレスポンスでは、配列の先頭要素に取得可能な
// 全作品数が格納され、2要素目以降に作品情報が格納されます。
type Response struct {
	AllCount int
	Novels   []Novel
}

// Novel は、なろう小説APIから返される作品情報です。
type Novel struct {
	// 基本情報
	Title   string `json:"title"`
	NCode   string `json:"ncode"`
	UserID  int    `json:"userid"`
	Writer  string `json:"writer"`
	Story   string `json:"story"`
	Keyword string `json:"keyword"`

	// ジャンル
	BigGenre BigGenre `json:"biggenre"`
	Genre    Genre    `json:"genre"`

	// 現在未使用で、常に空文字列です。
	OriginalWork string `json:"gensaku"`

	// 掲載日時
	//
	// 形式はYYYY-MM-DD HH:MM:SSです。
	GeneralFirstUp string `json:"general_firstup"`
	GeneralLastUp  string `json:"general_lastup"`

	// 作品状態
	//
	// NovelTypeはカスタムUnmarshalJSONで設定します。
	NovelType NovelType  `json:"-"`
	End       EndStatus  `json:"end"`
	IsStopped BinaryFlag `json:"isstop"`

	// 作品量
	EpisodeCount int `json:"general_all_no"`
	Length       int `json:"length"`
	ReadingTime  int `json:"time"`

	// 含有要素
	IsR15    BinaryFlag `json:"isr15"`
	IsBL     BinaryFlag `json:"isbl"`
	IsGL     BinaryFlag `json:"isgl"`
	IsCruel  BinaryFlag `json:"iszankoku"`
	IsTensei BinaryFlag `json:"istensei"`
	IsTenni  BinaryFlag `json:"istenni"`

	// ポイント
	GlobalPoint  int `json:"global_point"`
	DailyPoint   int `json:"daily_point"`
	WeeklyPoint  int `json:"weekly_point"`
	MonthlyPoint int `json:"monthly_point"`
	QuarterPoint int `json:"quarter_point"`
	YearlyPoint  int `json:"yearly_point"`

	// 反応・評価
	FavoriteNovelCount int `json:"fav_novel_cnt"`
	ImpressionCount    int `json:"impression_cnt"`
	ReviewCount        int `json:"review_cnt"`
	EvaluationPoint    int `json:"all_point"`
	EvaluatorCount     int `json:"all_hyoka_cnt"`

	// 作品内容の統計
	IllustrationCount int `json:"sasie_cnt"`
	DialogueRate      int `json:"kaiwaritu"`

	// API内部の更新日時
	NovelUpdatedAt string `json:"novelupdated_at"`
	UpdatedAt      string `json:"updated_at"`

	// opt=weeklyを指定した場合のみ出力されます。
	WeeklyUnique *int `json:"weekly_unique,omitempty"`
}

// IsShort は、作品が短編ならtrueを返します。
func (n Novel) IsShort() bool {
	return n.NovelType == NovelTypeShort
}

// IsSerial は、作品が連載作品ならtrueを返します。
func (n Novel) IsSerial() bool {
	return n.NovelType == NovelTypeSerial
}

// IsCompletedSerial は、作品が完結済みの連載作品ならtrueを返します。
func (n Novel) IsCompletedSerial() bool {
	return n.NovelType == NovelTypeSerial &&
		n.End == EndCompleted
}

// IsOngoingSerial は、作品が連載中ならtrueを返します。
func (n Novel) IsOngoingSerial() bool {
	return n.NovelType == NovelTypeSerial &&
		n.End == EndOngoing
}

// UnmarshalJSON は、なろうAPIの作品情報をNovelへデコードします。
//
// 通常のレスポンスでは作品タイプのキーがnovel_typeですが、
// ofパラメーターで作品タイプを指定した場合はnoveltypeになります。
// このメソッドでは両方のキーに対応します。
func (n *Novel) UnmarshalJSON(data []byte) error {
	type novelAlias Novel

	var raw struct {
		novelAlias

		NovelType       *NovelType `json:"novel_type"`
		NovelTypeWithOF *NovelType `json:"noveltype"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("decode novel: %w", err)
	}

	*n = Novel(raw.novelAlias)

	switch {
	case raw.NovelType != nil:
		n.NovelType = *raw.NovelType
	case raw.NovelTypeWithOF != nil:
		n.NovelType = *raw.NovelTypeWithOF
	}

	return nil
}

// UnmarshalJSON は、なろうAPI特有の配列形式をResponseへデコードします。
func (r *Response) UnmarshalJSON(data []byte) error {
	var elements []json.RawMessage

	if err := json.Unmarshal(data, &elements); err != nil {
		return fmt.Errorf("decode response array: %w", err)
	}

	if len(elements) == 0 {
		r.AllCount = 0
		r.Novels = nil
		return nil
	}

	var metadata struct {
		AllCount int `json:"allcount"`
	}

	if err := json.Unmarshal(elements[0], &metadata); err != nil {
		return fmt.Errorf("decode response metadata: %w", err)
	}

	novels := make([]Novel, 0, len(elements)-1)

	for index, element := range elements[1:] {
		if bytes.Equal(bytes.TrimSpace(element), []byte("null")) {
			continue
		}

		var novel Novel

		if err := json.Unmarshal(element, &novel); err != nil {
			return fmt.Errorf(
				"decode novel at index %d: %w",
				index+1,
				err,
			)
		}

		novels = append(novels, novel)
	}

	r.AllCount = metadata.AllCount
	r.Novels = novels

	return nil
}

// ParseTime は、なろうAPIの日時文字列を日本時間のtime.Timeへ変換します。
func ParseTime(value string) (time.Time, error) {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, fmt.Errorf("load Asia/Tokyo location: %w", err)
	}

	parsed, err := time.ParseInLocation(
		narouTimeLayout,
		value,
		location,
	)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"parse narou time %q: %w",
			value,
			err,
		)
	}

	return parsed, nil
}
