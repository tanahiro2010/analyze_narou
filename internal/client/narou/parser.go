package narou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

const narouTimeLayout = "2006-01-02 15:04:05"

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
//
// 配列の先頭要素には取得可能な全作品数が格納され、
// 2要素目以降には作品情報が格納されます。
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

	allCount, err := parseResponseMetadata(elements[0])
	if err != nil {
		return err
	}

	novels, err := parseNovels(elements[1:])
	if err != nil {
		return err
	}

	r.AllCount = allCount
	r.Novels = novels

	return nil
}

// parseResponseMetadata は、レスポンス先頭のメタデータを解析します。
func parseResponseMetadata(data json.RawMessage) (int, error) {
	var metadata struct {
		AllCount int `json:"allcount"`
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return 0, fmt.Errorf("decode response metadata: %w", err)
	}

	return metadata.AllCount, nil
}

// parseNovels は、レスポンス内の作品情報を解析します。
func parseNovels(elements []json.RawMessage) ([]Novel, error) {
	novels := make([]Novel, 0, len(elements))

	for index, element := range elements {
		if isJSONNull(element) {
			continue
		}

		var novel Novel

		if err := json.Unmarshal(element, &novel); err != nil {
			// APIレスポンス全体では先頭にメタデータがあるため、index+1とします。
			return nil, fmt.Errorf(
				"decode novel at response index %d: %w",
				index+1,
				err,
			)
		}

		novels = append(novels, novel)
	}

	return novels, nil
}

// isJSONNull は、JSON要素がnullかどうかを判定します。
func isJSONNull(data json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(data), []byte("null"))
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

func (r *RankingResult) UnmarshalJSON(data []byte) error {
	var items []RankingItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode ranking items: %w", err)
	}
	return nil
}
