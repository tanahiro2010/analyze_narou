package analytics

import "analyze_narou/internal/client/narou"

type GenreAnalyzeResult struct {
	TargetGenreName        string
	NovelCount             int
	BigGenreDistribution   []BigGenreCount
	GenreDistribution      []GenreCount
	TitleStoryAnalysis     TitleStoryAnalysis
	TagDistribution        []TagCount
	TagDistributionByGenre []GenreTagDistribution
	BookmarkAnalysis       BookmarkAnalysis
	EvaluationAnalysis     EvaluationAnalysis
	LengthAnalysis         LengthAnalysis
	PointAnalysis          PointAnalysis
	SerializationAnalysis  SerializationAnalysis
	DialogueAnalysis       DialogueAnalysis
	AIInsight              AIInsight
}

type AllAnalyzeResult struct {
	GenreResultCount      int
	NovelCount            int
	TagDistribution       []TagCount
	BookmarkAnalysis      BookmarkAnalysis
	EvaluationAnalysis    EvaluationAnalysis
	LengthAnalysis        LengthAnalysis
	PointAnalysis         PointAnalysis
	SerializationAnalysis SerializationAnalysis
	GenreSummaries        []GenreSummary
	WritingHints          []string
	AIInsight             AIInsight
}

type AIInsight struct {
	Summary           string
	TitleAndStory     string
	TagAndGenre       string
	ReaderSignal      string
	WritingAdvice     []string
	RecommendedTags   []string
	RecommendedTitles []TitleSuggestion
	CreativeTips      []CreativeTip
	Raw               string
	UnavailableReason string
}

type TitleSuggestion struct {
	Title     string
	Rationale string
}

type CreativeTip struct {
	Tip    string
	Source string
}

type BigGenreCount struct {
	BigGenre narou.BigGenre
	Count    int
	Rate     float64
}

type GenreCount struct {
	Genre narou.Genre
	Count int
	Rate  float64
}

type TitleStoryAnalysis struct {
	Title              TitleAnalysis
	Story              StoryAnalysis
	RepresentativeWork []NovelDigest
}

type TitleAnalysis struct {
	AverageLength             float64
	MinLength                 int
	MaxLength                 int
	LongTitleRate             float64
	QuestionOrExclamationRate float64
	BracketTitleRate          float64
}

type StoryAnalysis struct {
	AverageLength     float64
	MinLength         int
	MaxLength         int
	DepthDistribution StoryDepthDistribution
	CommonTerms       []TermCount
}

type StoryDepthDistribution struct {
	SetupOnly           int
	GoalOrConflict      int
	DevelopmentIncluded int
	EndingOrSpoiler     int
}

type TermCount struct {
	Term  string
	Count int
	Rate  float64
}

type TagCount struct {
	Tag   string
	Count int
	Rate  float64
}

type GenreTagDistribution struct {
	Genre narou.Genre
	Count int
	Tags  []TagCount
}

type BookmarkAnalysis struct {
	TotalBookmarks          int
	AverageBookmarks        float64
	BookmarkToEvaluatorRate float64
	BookmarkPointShare      float64
}

type EvaluationAnalysis struct {
	TotalEvaluationPoints     int
	TotalEvaluators           int
	AverageEvaluationPoint    float64
	AverageEvaluatorCount     float64
	AverageRatingPerEvaluator float64
}

type LengthAnalysis struct {
	TotalLength         int
	AverageLength       float64
	MedianLength        float64
	MinLength           int
	MaxLength           int
	TotalEpisodeCount   int
	AverageEpisodeCount float64
}

type PointAnalysis struct {
	TotalGlobalPoint   int
	AverageGlobalPoint float64
	TopGlobalPoint     []NovelDigest
}

type SerializationAnalysis struct {
	ShortCount           int
	SerialCount          int
	CompletedSerialCount int
	OngoingSerialCount   int
	StoppedCount         int
	CompletionRate       float64
	StoppedRate          float64
}

type DialogueAnalysis struct {
	AverageDialogueRate float64
}

type NovelDigest struct {
	NCode       string
	Title       string
	StoryDigest string
	GlobalPoint int
	Length      int
}

type GenreSummary struct {
	Genre               narou.Genre
	NovelCount          int
	TopTags             []TagCount
	AverageBookmarkRate float64
	AverageRating       float64
	AverageLength       float64
	AverageGlobalPoint  float64
}
