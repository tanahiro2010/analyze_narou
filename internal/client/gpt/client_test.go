package gpt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient(OpenAIConfig{
		ApiKey: "test-api-key",
		Model:  openai.GPT3Dot5Turbo,
	})

	if client == nil {
		t.Fatal("client is nil")
	}

	if client.client == nil {
		t.Fatal("openai client is nil")
	}

	if client.model != openai.GPT3Dot5Turbo {
		t.Fatalf("model = %q, want %q", client.model, openai.GPT3Dot5Turbo)
	}
}

func TestChatSendsPromptAndReturnsResponse(t *testing.T) {
	var gotRequest openai.ChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}

		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
			ID:    "chatcmpl-test",
			Model: gotRequest.Model,
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "response",
					},
				},
			},
		})
	}))
	defer server.Close()

	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = server.URL + "/v1"

	client := &OpenAIClient{
		client: openai.NewClientWithConfig(config),
		model:  "test-model",
	}

	responses, err := client.Chat([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}

	if gotRequest.Model != "test-model" {
		t.Fatalf("request model = %q, want test-model", gotRequest.Model)
	}

	if len(gotRequest.Messages) != 1 || gotRequest.Messages[0].Content != "hello" {
		t.Fatalf("unexpected request messages: %+v", gotRequest.Messages)
	}

	if len(responses) != 1 {
		t.Fatalf("len(responses) = %d, want 1", len(responses))
	}

	if responses[0].Choices[0].Message.Content != "response" {
		t.Fatalf("response content = %q, want response", responses[0].Choices[0].Message.Content)
	}
}

func TestChatReturnsWrappedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failed", http.StatusInternalServerError)
	}))
	defer server.Close()

	config := openai.DefaultConfig("test-api-key")
	config.BaseURL = server.URL + "/v1"

	client := &OpenAIClient{
		client: openai.NewClientWithConfig(config),
		model:  "test-model",
	}

	if _, err := client.Chat([]openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
	}); err == nil {
		t.Fatal("expected error")
	}
}
