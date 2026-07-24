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
		if field.Name == name && strings.Contains(field.Value, valuePart) {
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
