package e2e

import (
	"net/http"
	"testing"

	"github.com/yetiz-org/goth-scaffold/tests/e2e/testutils"
)

// TestHealthEndpoint verifies that the /api/v1/health endpoint returns 200.
// This is the baseline E2E test — if it fails, something is fundamentally broken.
func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	tc := testutils.NewTestContext(t)

	resp, err := tc.DoGet("/api/v1/health")
	if err != nil {
		t.Fatalf("GET /api/v1/health failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
