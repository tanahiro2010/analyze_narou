package analytics

import (
	"analyze_narou/internal/client/narou"
	"errors"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewAnalyzer(t *testing.T) {
	openAIClient := &fakeChatClient{}

	analyzer := NewAnalyzer(openAIClient)
	if analyzer == nil {
		t.Fatal("analyzer is nil")
	}
}

func TestGenreAnalyzeCalculatesRankingMetrics(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	result, err := analyzer.GenreAnalyze([]narou.Novel{
		{
			NCode:              "N1",
			Title:              "追放された魔法使いは辺境で成り上がる！",
			Story:              "追放された主人公が辺境で仲間を守るために戦う。そして王国の陰謀に巻き込まれていく。",
			Keyword:            "異世界 魔法 追放",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreHighFantasy,
			NovelType:          narou.NovelTypeSerial,
			End:                narou.EndOngoing,
			FavoriteNovelCount: 30,
			EvaluationPoint:    80,
			EvaluatorCount:     10,
			GlobalPoint:        140,
			Length:             100000,
			EpisodeCount:       20,
			DialogueRate:       40,
		},
		{
			NCode:              "N2",
			Title:              "【短編】魔女と最後の約束",
			Story:              "魔女と少年が最後に交わした約束の真相を描く。",
			Keyword:            "魔法 恋愛 短編",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreLowFantasy,
			NovelType:          narou.NovelTypeShort,
			End:                narou.EndCompleted,
			IsStopped:          narou.FlagDisabled,
			FavoriteNovelCount: 10,
			EvaluationPoint:    20,
			EvaluatorCount:     5,
			GlobalPoint:        40,
			Length:             20000,
			EpisodeCount:       1,
			DialogueRate:       20,
		},
	})
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	if result.NovelCount != 2 {
		t.Fatalf("NovelCount = %d, want 2", result.NovelCount)
	}

	if result.TitleStoryAnalysis.Title.AverageLength == 0 {
		t.Fatal("expected title analysis")
	}

	if result.TitleStoryAnalysis.Story.DepthDistribution.EndingOrSpoiler != 1 {
		t.Fatalf("EndingOrSpoiler = %d, want 1", result.TitleStoryAnalysis.Story.DepthDistribution.EndingOrSpoiler)
	}

	if result.TitleStoryAnalysis.Story.DepthDistribution.DevelopmentIncluded != 1 {
		t.Fatalf("DevelopmentIncluded = %d, want 1", result.TitleStoryAnalysis.Story.DepthDistribution.DevelopmentIncluded)
	}

	if len(result.TagDistribution) == 0 || result.TagDistribution[0].Tag != "魔法" || result.TagDistribution[0].Count != 2 {
		t.Fatalf("unexpected tag distribution: %+v", result.TagDistribution)
	}

	if len(result.TagDistributionByGenre) != 2 {
		t.Fatalf("len(TagDistributionByGenre) = %d, want 2", len(result.TagDistributionByGenre))
	}

	if result.BookmarkAnalysis.TotalBookmarks != 40 {
		t.Fatalf("TotalBookmarks = %d, want 40", result.BookmarkAnalysis.TotalBookmarks)
	}

	if result.BookmarkAnalysis.BookmarkToEvaluatorRate != float64(40)/15 {
		t.Fatalf("BookmarkToEvaluatorRate = %f, want %f", result.BookmarkAnalysis.BookmarkToEvaluatorRate, float64(40)/15)
	}

	if result.EvaluationAnalysis.AverageRatingPerEvaluator != float64(100)/15 {
		t.Fatalf("AverageRatingPerEvaluator = %f, want %f", result.EvaluationAnalysis.AverageRatingPerEvaluator, float64(100)/15)
	}

	if result.LengthAnalysis.AverageLength != 60000 {
		t.Fatalf("AverageLength = %f, want 60000", result.LengthAnalysis.AverageLength)
	}

	if result.SerializationAnalysis.ShortCount != 1 || result.SerializationAnalysis.OngoingSerialCount != 1 {
		t.Fatalf("unexpected serialization analysis: %+v", result.SerializationAnalysis)
	}

	if len(result.PointAnalysis.TopGlobalPoint) == 0 || result.PointAnalysis.TopGlobalPoint[0].NCode != "N1" {
		t.Fatalf("unexpected top novels: %+v", result.PointAnalysis.TopGlobalPoint)
	}

	if !strings.Contains(result.String(), "上位タグ") {
		t.Fatalf("String() = %q, want tag summary", result.String())
	}
}

func TestGenreAnalyzeHandlesEmptyInput(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	result, err := analyzer.GenreAnalyze(nil)
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	if result.NovelCount != 0 {
		t.Fatalf("NovelCount = %d, want 0", result.NovelCount)
	}

	if got := result.String(); got != "作品数: 0" {
		t.Fatalf("String() = %q, want empty summary", got)
	}
}

func TestAllAnalyzeAggregatesGenreResults(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	fantasy, err := analyzer.GenreAnalyze([]narou.Novel{
		{
			NCode:              "N1",
			Title:              "魔法使いの冒険",
			Story:              "主人公が仲間を守るために戦う。",
			Keyword:            "異世界 魔法",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreHighFantasy,
			NovelType:          narou.NovelTypeSerial,
			End:                narou.EndCompleted,
			FavoriteNovelCount: 20,
			EvaluationPoint:    40,
			EvaluatorCount:     5,
			GlobalPoint:        80,
			Length:             50000,
			EpisodeCount:       10,
		},
	})
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	romance, err := analyzer.GenreAnalyze([]narou.Novel{
		{
			NCode:              "N2",
			Title:              "恋の始まり",
			Story:              "二人の恋が始まる。",
			Keyword:            "恋愛 学園",
			BigGenre:           narou.BigGenreRomance,
			Genre:              narou.GenreRealRomance,
			NovelType:          narou.NovelTypeShort,
			End:                narou.EndCompleted,
			FavoriteNovelCount: 10,
			EvaluationPoint:    20,
			EvaluatorCount:     5,
			GlobalPoint:        40,
			Length:             10000,
			EpisodeCount:       1,
		},
	})
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	result, err := analyzer.AllAnalyze([]GenreAnalyzeResult{fantasy, romance})
	if err != nil {
		t.Fatalf("AllAnalyze returned error: %v", err)
	}

	if result.GenreResultCount != 2 {
		t.Fatalf("GenreResultCount = %d, want 2", result.GenreResultCount)
	}

	if result.NovelCount != 2 {
		t.Fatalf("NovelCount = %d, want 2", result.NovelCount)
	}

	if result.BookmarkAnalysis.TotalBookmarks != 30 {
		t.Fatalf("TotalBookmarks = %d, want 30", result.BookmarkAnalysis.TotalBookmarks)
	}

	if result.EvaluationAnalysis.AverageRatingPerEvaluator != 6 {
		t.Fatalf("AverageRatingPerEvaluator = %f, want 6", result.EvaluationAnalysis.AverageRatingPerEvaluator)
	}

	if len(result.TagDistribution) == 0 {
		t.Fatal("expected tag distribution")
	}

	if len(result.GenreSummaries) != 2 {
		t.Fatalf("len(GenreSummaries) = %d, want 2", len(result.GenreSummaries))
	}

	if len(result.WritingHints) == 0 {
		t.Fatal("expected writing hints")
	}

	if !strings.Contains(result.String(), "総作品数") {
		t.Fatalf("String() = %q, want total summary", result.String())
	}
}

func TestAllAnalyzeHandlesEmptyInput(t *testing.T) {
	analyzer := NewAnalyzer(nil)

	result, err := analyzer.AllAnalyze(nil)
	if err != nil {
		t.Fatalf("AllAnalyze returned error: %v", err)
	}

	if result.NovelCount != 0 {
		t.Fatalf("NovelCount = %d, want 0", result.NovelCount)
	}

	if got := result.String(); got != "作品数: 0" {
		t.Fatalf("String() = %q, want empty summary", got)
	}
}

func TestGenreAnalyzeAddsAIInsight(t *testing.T) {
	chatClient := &fakeChatClient{
		content: `{
			"summary":"上位作は導入と目的が明快です",
			"title_and_story":"長めのタイトルで売りを明示し、紹介文は対立まで書いています",
			"tag_and_genre":"魔法と追放が読者期待を作っています",
			"reader_signal":"ブックマーク比が高く、継続読書の訴求が強いです",
			"writing_advice":["タイトルに強みを入れる","紹介文で目的を早く出す"],
			"recommended_tags":["異世界","追放","魔法"],
			"recommended_titles":[
				{"title":"追放魔法使いは辺境で成り上がる","rationale":"長文タイトル率と追放タグの強さを踏まえています"},
				{"title":"弱小ギルドの魔法参謀","rationale":"目的・対立まで書く紹介文傾向に合わせています"}
			],
			"creative_tips":[
				{"tip":"序盤で主人公の欠落を見せる","source":"代表作品サンプルに主人公の尖りが早く出ています"},
				{"tip":"タグと紹介文の約束を揃える","source":"上位タグと紹介文の読者期待を接続しています"}
			]
		}`,
	}
	analyzer := NewAnalyzer(chatClient)

	result, err := analyzer.GenreAnalyze([]narou.Novel{
		{
			NCode:              "N1",
			Title:              "追放魔法使い",
			Story:              "主人公が仲間を守るために戦う。",
			Keyword:            "異世界 魔法 追放",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreHighFantasy,
			NovelType:          narou.NovelTypeSerial,
			End:                narou.EndOngoing,
			FavoriteNovelCount: 30,
			EvaluationPoint:    80,
			EvaluatorCount:     10,
			GlobalPoint:        140,
			Length:             100000,
			EpisodeCount:       20,
		},
	}, narou.BigGenreFantasy.String())
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	if chatClient.callCount != 1 {
		t.Fatalf("AI call count = %d, want 1", chatClient.callCount)
	}

	if !strings.Contains(chatClient.lastPrompt, "title_analysis") {
		t.Fatalf("prompt = %q, want title_analysis", chatClient.lastPrompt)
	}

	if !strings.Contains(chatClient.lastPrompt, `"target_genre_name": "ファンタジー"`) {
		t.Fatalf("prompt = %q, want target genre name", chatClient.lastPrompt)
	}

	if !strings.Contains(chatClient.lastPrompt, `"name": "ファンタジー"`) {
		t.Fatalf("prompt = %q, want named big genre distribution", chatClient.lastPrompt)
	}

	if !strings.Contains(chatClient.lastPrompt, `"name": "ハイファンタジー〔ファンタジー〕"`) {
		t.Fatalf("prompt = %q, want named genre distribution", chatClient.lastPrompt)
	}

	if result.TargetGenreName != "ファンタジー" {
		t.Fatalf("TargetGenreName = %q, want ファンタジー", result.TargetGenreName)
	}

	if result.AIInsight.Summary != "上位作は導入と目的が明快です" {
		t.Fatalf("AI summary = %q", result.AIInsight.Summary)
	}

	if len(result.AIInsight.WritingAdvice) != 2 {
		t.Fatalf("WritingAdvice = %+v, want 2 items", result.AIInsight.WritingAdvice)
	}

	if len(result.AIInsight.RecommendedTags) != 3 {
		t.Fatalf("RecommendedTags = %+v, want 3 items", result.AIInsight.RecommendedTags)
	}

	if len(result.AIInsight.RecommendedTitles) != 2 {
		t.Fatalf("RecommendedTitles = %+v, want 2 items", result.AIInsight.RecommendedTitles)
	}
	if result.AIInsight.RecommendedTitles[0].Rationale == "" {
		t.Fatalf("RecommendedTitles = %+v, want rationale", result.AIInsight.RecommendedTitles)
	}

	if len(result.AIInsight.CreativeTips) != 2 {
		t.Fatalf("CreativeTips = %+v, want 2 items", result.AIInsight.CreativeTips)
	}
	if result.AIInsight.CreativeTips[0].Source == "" {
		t.Fatalf("CreativeTips = %+v, want source", result.AIInsight.CreativeTips)
	}

	if !strings.Contains(result.String(), "AI要約") {
		t.Fatalf("String() = %q, want AI summary", result.String())
	}

	if !strings.Contains(result.String(), "AIおすすめタグ") {
		t.Fatalf("String() = %q, want recommended tags", result.String())
	}
}

func TestAllAnalyzeAddsAIInsight(t *testing.T) {
	chatClient := &fakeChatClient{
		content: `{"summary":"全体傾向","title_and_story":"タイトル傾向","tag_and_genre":"タグ傾向","reader_signal":"読者反応","writing_advice":["企画を絞る"],"recommended_tags":["恋愛","学園"],"recommended_titles":[{"title":"週末だけの契約恋人","rationale":"恋愛タグとタイトル傾向を踏まえています"}],"creative_tips":[{"tip":"一話目で関係性の火種を置く","source":"ジャンル別タグと紹介文傾向に基づきます"}]}`,
	}
	analyzer := NewAnalyzer(chatClient)

	genreResult := analyzeNovels([]narou.Novel{
		{
			NCode:              "N1",
			Title:              "魔法使いの冒険",
			Story:              "主人公が仲間を守るために戦う。",
			Keyword:            "異世界 魔法",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreHighFantasy,
			NovelType:          narou.NovelTypeSerial,
			End:                narou.EndCompleted,
			FavoriteNovelCount: 20,
			EvaluationPoint:    40,
			EvaluatorCount:     5,
			GlobalPoint:        80,
			Length:             50000,
			EpisodeCount:       10,
		},
	})

	result, err := analyzer.AllAnalyze([]GenreAnalyzeResult{genreResult})
	if err != nil {
		t.Fatalf("AllAnalyze returned error: %v", err)
	}

	if chatClient.callCount != 1 {
		t.Fatalf("AI call count = %d, want 1", chatClient.callCount)
	}

	if !strings.Contains(chatClient.lastPrompt, "all_rankings") {
		t.Fatalf("prompt = %q, want all_rankings", chatClient.lastPrompt)
	}

	if !strings.Contains(chatClient.lastPrompt, `"genre_summaries_with_names"`) {
		t.Fatalf("prompt = %q, want named genre summaries", chatClient.lastPrompt)
	}

	if !strings.Contains(chatClient.lastPrompt, `"name": "ハイファンタジー〔ファンタジー〕"`) {
		t.Fatalf("prompt = %q, want genre summary name", chatClient.lastPrompt)
	}

	if result.AIInsight.Summary != "全体傾向" {
		t.Fatalf("AI summary = %q", result.AIInsight.Summary)
	}
}

func TestGenreAnalyzeKeepsMetricsWhenAIError(t *testing.T) {
	analyzer := NewAnalyzer(&fakeChatClient{err: errors.New("temporary failure")})

	result, err := analyzer.GenreAnalyze([]narou.Novel{
		{
			NCode:              "N1",
			Title:              "魔法使いの冒険",
			Story:              "主人公が仲間を守るために戦う。",
			Keyword:            "異世界 魔法",
			BigGenre:           narou.BigGenreFantasy,
			Genre:              narou.GenreHighFantasy,
			NovelType:          narou.NovelTypeSerial,
			End:                narou.EndCompleted,
			FavoriteNovelCount: 20,
			EvaluationPoint:    40,
			EvaluatorCount:     5,
			GlobalPoint:        80,
			Length:             50000,
			EpisodeCount:       10,
		},
	})
	if err != nil {
		t.Fatalf("GenreAnalyze returned error: %v", err)
	}

	if result.NovelCount != 1 {
		t.Fatalf("NovelCount = %d, want 1", result.NovelCount)
	}

	if !strings.Contains(result.AIInsight.UnavailableReason, "temporary failure") {
		t.Fatalf("UnavailableReason = %q", result.AIInsight.UnavailableReason)
	}
}

type fakeChatClient struct {
	content    string
	err        error
	callCount  int
	lastPrompt string
}

func (f *fakeChatClient) Chat(prompts []openai.ChatCompletionMessage) ([]openai.ChatCompletionResponse, error) {
	f.callCount++
	if len(prompts) > 0 {
		f.lastPrompt = prompts[len(prompts)-1].Content
	}
	if f.err != nil {
		return nil, f.err
	}
	return []openai.ChatCompletionResponse{
		{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: f.content,
					},
				},
			},
		},
	}, nil
}
