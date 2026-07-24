package logger

import (
	"analyze_narou/internal/analytics"
	"analyze_narou/internal/client/discord"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewWebhookLogger(t *testing.T) {
	discordClient := discord.DiscordClient{}

	logger := NewWebhookLogger(discordClient)
	if logger == nil {
		t.Fatal("logger is nil")
	}
}

func TestLogSendsDiscordWebhookMessage(t *testing.T) {
	var gotMessage discord.WebhookMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotMessage); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})
	logger := NewWebhookLogger(*discordClient)

	if err := logger.Log("message"); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}

	if gotMessage.Username != "Narou Analyzer" {
		t.Fatalf("Username = %q, want Narou Analyzer", gotMessage.Username)
	}

	if gotMessage.Content != "message" {
		t.Fatalf("Content = %q, want message", gotMessage.Content)
	}
}

func TestLogSplitsLongDiscordMessages(t *testing.T) {
	var gotMessages []discord.WebhookMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var message discord.WebhookMessage
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		gotMessages = append(gotMessages, message)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})
	logger := NewWebhookLogger(*discordClient)

	if err := logger.Log(strings.Repeat("a", discordMessageLimit+10)); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}

	if len(gotMessages) != 2 {
		t.Fatalf("message count = %d, want 2", len(gotMessages))
	}

	for i, message := range gotMessages {
		if len([]rune(message.Content)) > discordMessageLimit {
			t.Fatalf("message[%d] length = %d, want <= %d", i, len([]rune(message.Content)), discordMessageLimit)
		}
	}
}

func TestLogReturnsErrorForDiscordErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})
	logger := NewWebhookLogger(*discordClient)

	err := logger.Log("message")
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "status 429") {
		t.Fatalf("error = %q, want status 429", err)
	}
}

func TestLogIgnoresEmptyMessage(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})
	logger := NewWebhookLogger(*discordClient)

	if err := logger.Log(" \n\t "); err != nil {
		t.Fatalf("Log returned error: %v", err)
	}

	if called {
		t.Fatal("expected empty message to skip webhook call")
	}
}

func TestGenreAnalyzeResultSendsFormattedSummary(t *testing.T) {
	gotMessage := sendAndCaptureMessage(t, func(logger *WebhookLogger) error {
		return logger.GenreAnalyzeResult(analytics.GenreAnalyzeResult{
			NovelCount: 1,
			TagDistribution: []analytics.TagCount{
				{Tag: "異世界", Count: 1},
			},
			AIInsight: analytics.AIInsight{Summary: "AI summary"},
		})
	})

	if gotMessage.Content != "ジャンル別ランキング分析" {
		t.Fatalf("Content = %q, want genre content", gotMessage.Content)
	}

	if len(gotMessage.Embeds) != 1 {
		t.Fatalf("len(Embeds) = %d, want 1", len(gotMessage.Embeds))
	}

	embed := gotMessage.Embeds[0]
	if embed.Title != "ジャンル別ランキング分析" {
		t.Fatalf("embed title = %q, want genre heading", embed.Title)
	}

	if !strings.Contains(embed.Description, "AI要約") || !strings.Contains(embed.Description, "AI summary") {
		t.Fatalf("description = %q, want AI summary", embed.Description)
	}

	if !embedHasField(embed, "上位タグ", "`異世界` 1件") {
		t.Fatalf("embed fields = %+v, want tag field", embed.Fields)
	}
}

func TestGenreAnalyzeResultSendsLongSampleAsDiscordEmbed(t *testing.T) {
	gotMessage := sendAndCaptureMessage(t, func(logger *WebhookLogger) error {
		return logger.GenreAnalyzeResult(longSampleGenreAnalyzeResult())
	})

	if gotMessage.Content != "ジャンル別ランキング分析" {
		t.Fatalf("Content = %q, want genre content", gotMessage.Content)
	}

	if len(gotMessage.Embeds) != 1 {
		t.Fatalf("len(Embeds) = %d, want 1", len(gotMessage.Embeds))
	}

	embed := gotMessage.Embeds[0]
	if embed.Title != "ジャンル別ランキング分析" {
		t.Fatalf("embed title = %q, want genre heading", embed.Title)
	}

	if len([]rune(embed.Description)) > discordEmbedFieldLimit {
		t.Fatalf("description length = %d, want <= %d", len([]rune(embed.Description)), discordEmbedFieldLimit)
	}

	if !strings.Contains(embed.Description, "VR系") {
		t.Fatalf("description = %q, want summary text", embed.Description)
	}

	wantFields := []string{
		"概要",
		"タイトル",
		"紹介文",
		"ブックマーク",
		"評価・文字数",
		"上位タグ",
		"AI タイトル・紹介分析",
		"AI タグ・ジャンル分析",
		"AI 読者反応分析",
		"AI 執筆アドバイス",
	}
	for _, name := range wantFields {
		if !embedHasField(embed, name, "") {
			t.Fatalf("embed fields = %+v, want field %q", embed.Fields, name)
		}
	}

	for _, field := range embed.Fields {
		if len([]rune(field.Value)) > discordEmbedFieldLimit {
			t.Fatalf("field %q length = %d, want <= %d", field.Name, len([]rune(field.Value)), discordEmbedFieldLimit)
		}
	}

	if !embedHasField(embed, "概要", "作品数: **100**") {
		t.Fatalf("embed fields = %+v, want novel count", embed.Fields)
	}
	if !embedHasField(embed, "上位タグ", "`VRMMO` 65件") {
		t.Fatalf("embed fields = %+v, want VRMMO tag", embed.Fields)
	}
	if !embedHasField(embed, "AI 執筆アドバイス", "入口の理解コスト") {
		t.Fatalf("embed fields = %+v, want writing advice", embed.Fields)
	}
}

func TestAllAnalyzeResultSendsFormattedSummary(t *testing.T) {
	gotMessage := sendAndCaptureMessage(t, func(logger *WebhookLogger) error {
		return logger.AllAnalyzeResult(analytics.AllAnalyzeResult{
			GenreResultCount: 1,
			NovelCount:       1,
			TagDistribution: []analytics.TagCount{
				{Tag: "恋愛", Count: 1},
			},
			WritingHints: []string{"紹介文で目的を出す"},
			AIInsight:    analytics.AIInsight{Summary: "All AI summary"},
		})
	})

	if gotMessage.Content != "全体ランキング分析" {
		t.Fatalf("Content = %q, want all content", gotMessage.Content)
	}

	if len(gotMessage.Embeds) != 1 {
		t.Fatalf("len(Embeds) = %d, want 1", len(gotMessage.Embeds))
	}

	embed := gotMessage.Embeds[0]
	if embed.Title != "全体ランキング分析" {
		t.Fatalf("embed title = %q, want all heading", embed.Title)
	}

	if !strings.Contains(embed.Description, "AI要約") || !strings.Contains(embed.Description, "All AI summary") {
		t.Fatalf("description = %q, want AI summary", embed.Description)
	}

	if !embedHasField(embed, "執筆ヒント", "紹介文で目的を出す") {
		t.Fatalf("embed fields = %+v, want writing hints field", embed.Fields)
	}
}

func sendAndCaptureContent(t *testing.T, send func(*WebhookLogger) error) string {
	t.Helper()

	message := sendAndCaptureMessage(t, send)
	return message.Content
}

func sendAndCaptureMessage(t *testing.T, send func(*WebhookLogger) error) discord.WebhookMessage {
	t.Helper()

	var gotMessage discord.WebhookMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotMessage); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	discordClient := discord.NewDiscordClient(discord.DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})
	logger := NewWebhookLogger(*discordClient)

	if err := send(logger); err != nil {
		t.Fatalf("send returned error: %v", err)
	}

	if gotMessage.Content == "" {
		t.Fatal("expected content")
	}

	return gotMessage
}

func embedHasField(embed discord.WebhookEmbed, name string, valuePart string) bool {
	for _, field := range embed.Fields {
		if field.Name == name && (valuePart == "" || strings.Contains(field.Value, valuePart)) {
			return true
		}
	}

	return false
}

func TestSplitDiscordMessagePrefersNewline(t *testing.T) {
	message := fmt.Sprintf("%s\n%s", strings.Repeat("a", 10), strings.Repeat("b", 10))

	chunks := splitDiscordMessage(message, 12)

	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}

	if chunks[0] != strings.Repeat("a", 10) {
		t.Fatalf("chunks[0] = %q", chunks[0])
	}

	if chunks[1] != strings.Repeat("b", 10) {
		t.Fatalf("chunks[1] = %q", chunks[1])
	}
}

func longSampleGenreAnalyzeResult() analytics.GenreAnalyzeResult {
	return analytics.GenreAnalyzeResult{
		NovelCount: 100,
		TitleStoryAnalysis: analytics.TitleStoryAnalysis{
			Title: analytics.TitleAnalysis{
				AverageLength:             32.4,
				LongTitleRate:             0.46,
				QuestionOrExclamationRate: 0.21,
				BracketTitleRate:          0.22,
			},
			Story: analytics.StoryAnalysis{
				AverageLength: 345.2,
				DepthDistribution: analytics.StoryDepthDistribution{
					SetupOnly:           27,
					GoalOrConflict:      22,
					DevelopmentIncluded: 50,
					EndingOrSpoiler:     1,
				},
			},
		},
		BookmarkAnalysis: analytics.BookmarkAnalysis{
			AverageBookmarks:        15417.3,
			BookmarkToEvaluatorRate: 4.76,
			BookmarkPointShare:      0.502,
		},
		EvaluationAnalysis: analytics.EvaluationAnalysis{
			AverageEvaluationPoint:    30587.5,
			AverageRatingPerEvaluator: 9.45,
		},
		LengthAnalysis: analytics.LengthAnalysis{
			AverageLength: 1201394,
		},
		TagDistribution: []analytics.TagCount{
			{Tag: "VRMMO", Count: 65},
			{Tag: "残酷な描写あり", Count: 59},
			{Tag: "R15", Count: 55},
			{Tag: "男主人公", Count: 44},
			{Tag: "ゲーム", Count: 24},
			{Tag: "女主人公", Count: 24},
			{Tag: "ほのぼの", Count: 22},
			{Tag: "冒険", Count: 22},
			{Tag: "未来", Count: 20},
			{Tag: "ＢＷＫ大賞１", Count: 20},
		},
		AIInsight: analytics.AIInsight{
			Summary:       "集計上はジャンル401が66%で、タグでもVRMMOがほぼ必須級（401内で約97%）となっており、このランキングはVR系（ゲーム/オンライン）中心の市場に寄っています。上位例（シャングリラ・フロンティア、痛いのは嫌なので…等）も「ゲーム前提の明確な入口」「主人公の尖った志向（クソゲー愛、防御極振り等）」「長期連載で積み上がる楽しさ」を示す紹介になっていました。ポイント構造は“ブクマが半分”で、短期爆発より継続読者の積み上げで上がりやすい傾向が読み取れます。",
			TitleAndStory: "タイトルは平均約32.4文字で長め、長タイトル率46%・カッコ付き22%・疑問符/感嘆符21%と、「一文でコンセプトを説明する」「装置（VR/宇宙船/最強装備など）を前面に出す」傾向が見られます。上位サンプルでも“何をする話か”がタイトルと冒頭数行で把握できます。紹介文（あらすじ）は平均約345文字で、深さ分布は『展開まで含む』が50件、『導入のみ』27件、『目的/対立まで』22件、『オチ/ネタバレ級』1件なので、読み手が最初に知りたいのは「導入＋何が面白さの軸か（主人公の目的、ゲーム/世界の特徴、基本の勝ち筋）」あたりまでで、結末を語りすぎない紹介が主流です。共通語にOnline/オンライン、プレイヤが多いことからも、紹介文内で“その作品のオンライン/ゲーム文脈”を早めに立ち上げるのが一般的です。なお「https」「書籍化」が一定数出ているため、外部展開・実績の告知を冒頭付近に置く例も混在しますが、ランキング全体の共通必須要素とまでは断定しにくいです。",
			TagAndGenre:   "ジャンル別に見ると、401（66%）はVRMMOが圧倒的で、次にR15/残酷な描写あり（各約61%）が目立ちます。これは「ゲーム要素＋一定の緊張感（危険/戦闘/痛み/デス等）」を期待されやすい可能性を示します。また401では「掲示板（約29%）」「ほのぼの（約30%）」「冒険（約27%）」も一定数あり、攻略・成長の合間に“コミュニティ反応/日常感”を混ぜる型が支持されている可能性があります。402（19%）は未来/スペースオペラ/ロボットが強く（各約68〜74%）、ここでもR15/残酷が過半で、戦闘や軍事・政治など硬派寄りの期待が生じやすい構造です。403（10%）はSF/人工知能/近未来が各30%程度で分散、404（5%）は現代×シリアス/ダーク寄り（残酷80%、シリアス60%）で少数派ながら尖った需要があるように見えます。全体上位タグにも「男主人公44%」「女主人公24%」が並ぶため、性別で極端に片寄るというより“題材（VR/未来）に合う主人公像”をタグで明示することが探索上は有利になりそうです。",
			ReaderSignal:  "平均評価（AverageRatingPerEvaluator）は約9.45と高めで、評価を付ける層の満足度は比較的高い傾向が示唆されます。一方でブックマーク数/評価者数（BookmarkToEvaluatorRate）が約4.76、かつ総合ポイントのうちブックマーク由来推定が約50.2%（BookmarkPointShare）なので、「評価を入れる」より「追いかける（ブクマする）」行動が強く、連載追従型の読まれ方が中心になりやすい可能性があります。文字数は平均約120万字・中央値約52万字、話数平均約317話で長期連載が主流（連載92/100、完結率約14.1%）なので、読者は“長く遊べる/読み続けられる”作品に反応しやすい構造が見えます。会話率平均約38%は、説明一辺倒よりテンポの良い掛け合い・状況進行を好む読者が一定数いる示唆になります。",
			WritingAdvice: []string{
				"入口の理解コストを下げる：タイトルと紹介文の冒頭1〜3行で「舞台（VRMMO/宇宙国家など）」「主人公の尖り（クソゲー愛、極振り等）」「何を積み上げる物語か（攻略/成長/傭兵稼業など）」をセットで提示すると、このランキング帯の読者が想定しやすいです（長タイトル率46%、紹介文は展開まで触れる例が最多）。",
				"“追いかけたくなる連載設計”を優先する：ブクマ由来推定が約50%で、ブクマ/評価者比も約4.76のため、単発の高評価より継続追従の積み上げが重要になりやすいです。毎話の小さな達成（新スキル獲得、レアドロ、勢力関係の前進など）を入れて「次も読む理由」を残す設計が相性良い可能性があります。",
				"会話と状況進行でテンポを作る：会話率平均約38%という集計から、説明だけで押すより、掛け合いで情報提示・キャラ立て・次の目標設定を回す作りが受け入れられやすい示唆があります。特にVR/未来題材は用語説明が増えがちなので、会話・行動の中で小出しにすると離脱を抑えやすいです。",
				"タグは“必須期待”と“差別化”を分けて設計する：401ならVRMMO/ゲームに加え、トーン（ほのぼの/シリアス）、装置（掲示板、近未来）、注意（R15/残酷）を整理して付けると、検索流入とミスマッチ低減の両方に寄与しやすいです。402なら未来/スペースオペラ/ロボット＋ミリタリー/ギャグ等で読み味を補足するのが自然です（ジャンル別タグ分布が比較的はっきりしているため）。",
				"紹介文は『導入＋目的/面白さの軸＋最初の一歩』までで止める：深さ分布では“展開まで”が最多ですが、“結末まで”はほぼ無いので、ネタバレよりも「何が読めるのか」を具体化する方が主流です。例：主人公の制約（痛いのが嫌、悪徳領主を目指す等）→世界/システムの特徴→最初の目標（攻略、稼ぐ、仲間づくり）という順で、読み手がブクマ判断しやすい形にするのが無難です。",
			},
		},
	}
}

func TestWebhookLoggerMethodsReturnNil(t *testing.T) {
	gotMessage := sendAndCaptureMessage(t, func(logger *WebhookLogger) error {
		return logger.GenreAnalyzeResult(analytics.GenreAnalyzeResult{})
	})

	if len(gotMessage.Embeds) != 1 || !strings.Contains(gotMessage.Embeds[0].Description, "作品数: 0") {
		t.Fatalf("message = %+v, want empty result summary", gotMessage)
	}

	gotMessage = sendAndCaptureMessage(t, func(logger *WebhookLogger) error {
		return logger.AllAnalyzeResult(analytics.AllAnalyzeResult{})
	})

	if len(gotMessage.Embeds) != 1 || !strings.Contains(gotMessage.Embeds[0].Description, "作品数: 0") {
		t.Fatalf("message = %+v, want empty result summary", gotMessage)
	}
}
