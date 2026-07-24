package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordConfig struct {
	WebhookURL string
	Timeout    time.Duration
}

type DiscordClient struct {
	webhookURL string
	client     *http.Client
}

type WebhookMessage struct {
	Username  string         `json:"username"`
	AvaterURL string         `json:"avater_url"`
	Content   string         `json:"content"`
	Embeds    []WebhookEmbed `json:"embeds,omitempty"`
}

type WebhookEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []WebhookEmbedField `json:"fields,omitempty"`
}

type WebhookEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func NewDiscordClient(config DiscordConfig) *DiscordClient {
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
