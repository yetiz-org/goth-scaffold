package googleauth

import (
	"strings"
	"testing"
)

func TestNewClientSetsCredentials(t *testing.T) {
	c := NewClient("my-client-id", "my-client-secret")
	if c.ClientId != "my-client-id" {
		t.Errorf("ClientId = %q, want %q", c.ClientId, "my-client-id")
	}
	if c.ClientSecret != "my-client-secret" {
		t.Errorf("ClientSecret = %q, want %q", c.ClientSecret, "my-client-secret")
	}
}

func TestGenerateOAuthUrlContainsClientId(t *testing.T) {
	c := NewClient("test-client-id", "test-secret")
	oauthUrl := c.GenerateOAuthUrl("https://example.com/callback", "random-state")

	if !strings.Contains(oauthUrl, "test-client-id") {
		t.Errorf("expected oauth URL to contain client_id, got: %s", oauthUrl)
	}
	if !strings.Contains(oauthUrl, "accounts.google.com") {
		t.Errorf("expected oauth URL to use Google, got: %s", oauthUrl)
	}
	if !strings.Contains(oauthUrl, "random-state") {
		t.Errorf("expected oauth URL to contain state, got: %s", oauthUrl)
	}
	if !strings.Contains(oauthUrl, "https%3A%2F%2Fexample.com%2Fcallback") {
		t.Errorf("expected oauth URL to contain encoded redirect_uri, got: %s", oauthUrl)
	}
}

func TestGenerateOAuthUrlContainsScope(t *testing.T) {
	c := NewClient("cid", "csecret")
	oauthUrl := c.GenerateOAuthUrl("https://app.example.com/callback", "state123")

	if !strings.Contains(oauthUrl, "openid") {
		t.Errorf("expected oauth URL to include openid scope, got: %s", oauthUrl)
	}
}
