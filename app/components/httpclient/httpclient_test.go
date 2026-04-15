package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRequestReturnsNilOnInvalidURL(t *testing.T) {
	req := NewRequest("GET", "://invalid", nil)
	if req != nil {
		t.Fatal("expected nil for invalid URL, got non-nil request")
	}
}

func TestNewRequestReturnsRequestOnValidURL(t *testing.T) {
	req := NewRequest("GET", "http://example.com/path", nil)
	if req == nil {
		t.Fatal("expected non-nil request for valid URL")
	}
	if req.Method != "GET" {
		t.Errorf("expected method GET, got %s", req.Method)
	}
}

func TestDoAndLogReturnsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	req := NewRequest("GET", server.URL+"/test", nil)
	if req == nil {
		t.Fatal("NewRequest returned nil")
	}

	resp, err := DoAndLog(req)
	if err != nil {
		t.Fatalf("DoAndLog() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Errorf("unexpected response body: %s", string(body))
	}
}

func TestDoReturnsResponseWithoutLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	req := NewRequest("POST", server.URL+"/create", strings.NewReader(`{}`))
	if req == nil {
		t.Fatal("NewRequest returned nil")
	}

	resp, err := Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}
