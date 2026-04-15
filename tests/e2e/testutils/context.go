package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestContext holds the state for a single E2E test case.
// Create one per test with NewTestContext; call t.Parallel() first.
type TestContext struct {
	T       *testing.T
	BaseURL string
	Client  *http.Client
}

// NewTestContext constructs a TestContext for the given test.
func NewTestContext(t *testing.T) *TestContext {
	t.Helper()

	return &TestContext{
		T:       t,
		BaseURL: GetBaseURL(),
		Client:  &http.Client{},
	}
}

// NewClient returns a fresh http.Client that does NOT follow redirects.
func (tc *TestContext) NewClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// DoGet performs a GET request to path (relative to BaseURL).
func (tc *TestContext) DoGet(path string) (*http.Response, error) {
	tc.T.Helper()

	return tc.Client.Get(tc.BaseURL + path)
}

// DoPostJSON performs a POST request with a JSON body.
func (tc *TestContext) DoPostJSON(path, body string) (*http.Response, error) {
	tc.T.Helper()

	req, err := http.NewRequest(http.MethodPost, tc.BaseURL+path, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return tc.Client.Do(req)
}

// ReadJSONBody reads and decodes a JSON response body into dest.
func ReadJSONBody(t *testing.T, resp *http.Response, dest any) {
	t.Helper()

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadJSONBody: read error: %v", err)
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(dest); err != nil {
		t.Fatalf("ReadJSONBody: decode error: %v (body: %s)", err, string(body))
	}
}
