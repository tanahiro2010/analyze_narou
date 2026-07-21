package gpt

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	ApiKey string
	Model  openai.Model
}
type OpenAIClient struct {
	client *openai.Client
	model  openai.Model
}

func NewOpenAIClient(config OpenAIConfig) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(config.ApiKey),
		model:  config.Model,
	}
}

func (c *OpenAIClient) Chat(prompts []openai.ChatCompletionMessage) ([]openai.ChatCompletionResponse, error) {
	res, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.model.ID,
			Messages: prompts,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	return []openai.ChatCompletionResponse{res}, nil
}
