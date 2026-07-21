package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordClientConfig struct {
	WebhookURL string
	Timeout    time.Duration
}

type DiscordClient struct {
	webhookURL string
	client     *http.Client
}

type WebhookMessage struct {
	Username  string `json:"username"`
	AvaterURL string `json:"avater_url"`
	Content   string `json:"content"`
}

func NewDiscordClient(config DiscordClientConfig) *DiscordClient {
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}
	return &DiscordClient{
		webhookURL: config.WebhookURL,
		client:     httpClient,
	}
}

func (c *DiscordClient) SendMessage(message WebhookMessage) (resp *http.Response, err error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.client.Post(c.webhookURL, "application/json", bytes.NewBuffer(payload))
}
