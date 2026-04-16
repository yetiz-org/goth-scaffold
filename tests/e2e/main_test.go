/*
Package e2e is the entry point for end-to-end tests.

It manages the full test-environment lifecycle:
 1. Load test configuration (config path, base URL)
 2. Build (if needed) and start the application binary
 3. Wait until the /api/v1/health endpoint responds with 200
 4. Execute all test functions
 5. Stop the server and clean up resources

All test functions MUST call t.Parallel() and obtain a *testutils.TestContext via
testutils.NewTestContext(t).

Run with:

	go test -v -count=1 ./tests/e2e/...

Set SCAFFOLD_E2E_BINARY to skip the build step and use an existing binary.
Set TEST_BASE_URL to skip starting a server and point at an already-running instance.
*/
package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yetiz-org/goth-scaffold/app"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/tests/e2e/testutils"
)

// TestMain manages the test environment lifecycle.
func TestMain(m *testing.M) {
	var exitCode int

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n  Panic occurred: %v\n", r)
			testutils.StopTestServer()
			os.Exit(1)
		}

		fmt.Printf("\nShutting down server...\n")
		testutils.StopTestServer()
		os.Exit(exitCode)
	}()

	// 1. Load test configuration
	testConfig := testutils.GetTestConfig()

	// Non-CI skip: config file not found — run unit tests only (make local-env-setup to enable e2e).
	if testConfig.ConfigPath == "" && os.Getenv("TEST_BASE_URL") == "" {
		fmt.Printf("  Skipping E2E tests: config not found — run: make local-env-setup && make local-db-seed\n")
		exitCode = 0
		return
	}

	fmt.Printf("E2E Test Configuration:\n")
	fmt.Printf("  - Config Path: %s\n", testConfig.ConfigPath)
	fmt.Printf("  - Base URL:    %s\n", testConfig.BaseURL)
	fmt.Printf("  - Is CI:       %v\n", testConfig.IsCI)

	// 2. Parse config so conf.Config() is populated
	absConfigPath := testConfig.ConfigPath
	if !strings.HasPrefix(absConfigPath, "/") {
		wd, _ := os.Getwd()
		absConfigPath = filepath.Join(wd, "../..", testConfig.ConfigPath)
	}

	os.Args = []string{"scaffold", "-c", absConfigPath, "-m", "default"}
	app.FlagParse()

	// Resolve secret path to absolute so connectors can find secrets from the test binary.
	// Normally set by daemon 01_setup_environment; here we replicate that for the test process.
	secretPath := conf.Config().DataStore.SecretPath
	if !filepath.IsAbs(secretPath) {
		secretPath = filepath.Join(testutils.GetProjectRoot(), secretPath)
	}
	os.Setenv("GOTH_SECRET_PATH", secretPath)

	fmt.Printf("  - Loaded config: %s\n", conf.ConfigPath)

	// 3. Skip server start if TEST_BASE_URL is already set (external server)
	if os.Getenv("TEST_BASE_URL") == "" {
		phaseStart := time.Now()
		fmt.Printf("\nStarting server...\n")

		if err := testutils.StartTestServer(absConfigPath); err != nil {
			fmt.Printf("  Failed to start server: %v\n", err)
			if testConfig.IsCI {
				exitCode = 1
			} else {
				fmt.Printf("  Skipping E2E tests (non-CI: server not available)\n")
				exitCode = 0
			}
			return
		}

		// 4. Wait for server to be ready
		baseURL := testutils.GetBaseURL()
		fmt.Printf("Waiting for server at %s...\n", baseURL)

		if err := testutils.WaitForServer(baseURL, 30*time.Second); err != nil {
			fmt.Printf("  Server not ready: %v\n", err)

			if logPath := testutils.GetServerLogPath(); logPath != "" {
				fmt.Printf("  Server log: %s\n", logPath)
			}

			if testConfig.IsCI {
				exitCode = 1
			} else {
				fmt.Printf("  Skipping E2E tests (non-CI: services not running — start with: make env-up)\n")
				exitCode = 0
			}
			return
		}

		fmt.Printf("  Server ready (took %v)\n\n", time.Since(phaseStart))
	}

	// 5. Execute tests
	testStart := time.Now()
	exitCode = m.Run()
	fmt.Printf("\n  All tests completed (took %v)\n", time.Since(testStart))
}
