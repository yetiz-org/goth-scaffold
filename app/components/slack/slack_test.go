package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientSetsWebhook(t *testing.T) {
	c := NewClient("https://hooks.slack.com/test")
	if c.Webhook != "https://hooks.slack.com/test" {
		t.Errorf("Webhook = %q, want %q", c.Webhook, "https://hooks.slack.com/test")
	}
}

func TestSendReturnsErrWebhookNotConfiguredWhenEmpty(t *testing.T) {
	c := NewClient("")
	err := c.Send("hello")
	if err == nil {
		t.Fatal("expected error for empty webhook, got nil")
	}
	if err != ErrWebhookNotConfigured {
		t.Errorf("expected ErrWebhookNotConfigured, got %v", err)
	}
}

func TestSendPostsPayloadToWebhook(t *testing.T) {
	var received Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL)
	err := c.Send("hello world")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if received.Text != "hello world" {
		t.Errorf("received text = %q, want %q", received.Text, "hello world")
	}
}

func TestSendWithContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	c := NewClient(server.URL)
	err := c.SendWithContext(ctx, "test")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestHTTPErrorMessage(t *testing.T) {
	e := &HTTPError{StatusCode: 400, Body: "bad request"}
	msg := e.Error()
	if msg != "slack webhook returned status 400: bad request" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestSendReturnsHTTPErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := NewClient(server.URL)
	err := c.Send("test")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", httpErr.StatusCode)
	}
}
