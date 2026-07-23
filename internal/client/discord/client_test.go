package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDiscordClient(t *testing.T) {
	client := NewDiscordClient(DiscordConfig{
		WebhookURL: "https://example.com/webhook",
		Timeout:    3 * time.Second,
	})

	if client.webhookURL != "https://example.com/webhook" {
		t.Fatalf("webhookURL = %q, want configured URL", client.webhookURL)
	}

	if client.client.Timeout != 3*time.Second {
		t.Fatalf("timeout = %s, want 3s", client.client.Timeout)
	}
}

func TestSendMessagePostsJSONPayload(t *testing.T) {
	var gotContentType string
	var gotMessage WebhookMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotMessage); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewDiscordClient(DiscordConfig{
		WebhookURL: server.URL,
		Timeout:    time.Second,
	})

	resp, err := client.SendMessage(WebhookMessage{
		Username:  "bot",
		AvaterURL: "https://example.com/avatar.png",
		Content:   "hello",
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}

	if gotMessage.Username != "bot" || gotMessage.AvaterURL != "https://example.com/avatar.png" || gotMessage.Content != "hello" {
		t.Fatalf("unexpected message: %+v", gotMessage)
	}
}
