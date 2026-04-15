package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrWebhookNotConfigured = errors.New("slack webhook not configured")

const defaultTimeout = 10 * time.Second

// HTTPError is returned when Slack webhook responds with a non-2xx status.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("slack webhook returned status %d: %s", e.StatusCode, e.Body)
}

// Payload is the Slack incoming webhook message body.
type Payload struct {
	Text string `json:"text,omitempty"`
}

// Client sends messages to a Slack incoming webhook.
type Client struct {
	Webhook string
}

// NewClient creates a new Slack Client with the given webhook URL.
func NewClient(webhook string) *Client {
	return &Client{Webhook: webhook}
}

// Send sends a text message using the default timeout.
func (c *Client) Send(text string) error {
	return c.SendWithTimeout(text, defaultTimeout)
}

// SendWithTimeout sends a text message with a custom timeout.
func (c *Client) SendWithTimeout(text string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.SendWithContext(ctx, text)
}

// SendWithContext sends a text message using the provided context.
func (c *Client) SendWithContext(ctx context.Context, text string) error {
	return sendText(ctx, c.Webhook, text)
}

func sendText(ctx context.Context, webhookURL string, text string) error {
	if webhookURL == "" {
		return ErrWebhookNotConfigured
	}

	requestBody, err := json.Marshal(&Payload{Text: text})
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return nil
}
