package gpt

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	ApiKey  string
	BaseURL string
	Model   string
}
type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(config OpenAIConfig) *OpenAIClient {
	openAIConfig := openai.DefaultConfig(config.ApiKey)
	if config.BaseURL != "" {
		openAIConfig.BaseURL = config.BaseURL
	}

	return &OpenAIClient{
		client: openai.NewClientWithConfig(openAIConfig),
		model:  config.Model,
	}
}

func (c *OpenAIClient) Chat(prompts []openai.ChatCompletionMessage) ([]openai.ChatCompletionResponse, error) {
	res, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: prompts,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	return []openai.ChatCompletionResponse{res}, nil
}
