package testutils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TestConfig holds configuration for the E2E test run.
type TestConfig struct {
	BaseURL    string
	ConfigPath string
	IsCI       bool
}

var (
	testConfigInstance   *TestConfig
	runtimeConfigCleanup func()
)

// GetTestConfig returns test configuration.
// In CI it expects evaluate/config.yaml.ci; locally it uses evaluate/config.yaml.local.
func GetTestConfig() *TestConfig {
	if testConfigInstance != nil {
		return testConfigInstance
	}

	isCI := os.Getenv("CI") == "true" || os.Getenv("GITLAB_CI") == "true"

	var configPath string
	if isCI {
		configPath = "evaluate/config.yaml.ci"
	} else {
		configPath = "evaluate/config.yaml.local"
	}

	runtimeConfigPath := configPath
	baseURL := os.Getenv("TEST_BASE_URL")

	if baseURL == "" {
		absConfigPath := configPath
		if absConfigPath[0] != '/' {
			absConfigPath = filepath.Join(GetProjectRoot(), configPath)
		}

		// Non-CI: skip gracefully when the local config file is absent.
		// Callers should check TestConfig.ConfigPath == "" and skip rather than fail.
		if !isCI {
			if _, statErr := os.Stat(absConfigPath); os.IsNotExist(statErr) {
				testConfigInstance = &TestConfig{ConfigPath: "", BaseURL: "", IsCI: false}
				return testConfigInstance
			}
		}

		var prepareErr error
		runtimeConfigPath, baseURL, _, runtimeConfigCleanup, prepareErr = PrepareRuntimeConfig(absConfigPath)
		if prepareErr != nil {
			if !isCI {
				// Still non-fatal locally — skip rather than crash.
				testConfigInstance = &TestConfig{ConfigPath: "", BaseURL: "", IsCI: false}
				return testConfigInstance
			}

			panic(fmt.Sprintf("failed to prepare e2e runtime config: %v", prepareErr))
		}
	}

	testConfigInstance = &TestConfig{
		ConfigPath: runtimeConfigPath,
		BaseURL:    baseURL,
		IsCI:       isCI,
	}

	return testConfigInstance
}

// GetBaseURL returns the base URL for E2E requests.
func GetBaseURL() string {
	return GetTestConfig().BaseURL
}

// GetProjectRoot returns the repository root (two levels up from tests/e2e/).
func GetProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}

	return filepath.Join(wd, "../..")
}

// CleanupRuntimeConfig removes the temporary config file created by GetTestConfig.
func CleanupRuntimeConfig() {
	if runtimeConfigCleanup != nil {
		runtimeConfigCleanup()
		runtimeConfigCleanup = nil
	}
}

// PrepareRuntimeConfig reads baseConfigPath, rewrites the app.port to a free port,
// writes a temp file, and returns (runtimePath, baseURL, port, cleanup, error).
func PrepareRuntimeConfig(baseConfigPath string) (string, string, int, func(), error) {
	selectedPort, err := findAvailablePort()
	if err != nil {
		return "", "", 0, nil, err
	}

	content, err := os.ReadFile(baseConfigPath)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to read base config: %w", err)
	}

	var cfg map[string]any
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to parse base config yaml: %w", err)
	}

	app, ok := cfg["app"].(map[string]any)
	if !ok {
		return "", "", 0, nil, fmt.Errorf("invalid config: app section not found or malformed")
	}

	app["port"] = selectedPort

	runtimeContent, err := yaml.Marshal(cfg)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to marshal runtime config yaml: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "scaffold-e2e-config-")
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to create runtime config temp dir: %w", err)
	}

	runtimeConfigPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(runtimeConfigPath, runtimeContent, 0644); err != nil {
		os.RemoveAll(tmpDir)
		return "", "", 0, nil, fmt.Errorf("failed to write runtime config file: %w", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Printf("Warning: failed to remove temp config dir %s: %v\n", tmpDir, err)
		}
	}

	runtimeBaseURL := fmt.Sprintf("http://localhost:%d", selectedPort)

	return runtimeConfigPath, runtimeBaseURL, selectedPort, cleanup, nil
}

func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to find available tcp port: %w", err)
	}

	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("failed to resolve tcp listener address")
	}

	return addr.Port, nil
}
