package testutils

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	serverCmd     *exec.Cmd
	serverLogFile *os.File
	serverLogPath string
	serverBinPath string
)

// StartTestServer builds (if needed) and starts the application binary with the
// given config file.  Use StopTestServer to shut it down after tests complete.
func StartTestServer(configPath string) error {
	projectRoot := GetProjectRoot()

	logDir := filepath.Join(projectRoot, "alloc", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	logPath := filepath.Join(logDir, fmt.Sprintf("scaffold-e2e-%s.log", timestamp))
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	serverLogFile = logFile
	serverLogPath = logPath

	fmt.Printf("  - Server log: %s\n", logPath)

	binaryPath, err := resolveServerBinary(projectRoot)
	if err != nil {
		logFile.Close()
		return err
	}

	serverCmd = exec.Command(binaryPath, "-c", configPath, "-m", "default")
	serverCmd.Dir = projectRoot
	serverCmd.Env = append(os.Environ(), "APP_DEBUG=true")
	serverCmd.Stdout = logFile
	serverCmd.Stderr = logFile
	serverCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := serverCmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start server: %w", err)
	}

	fmt.Printf("  - Server PID: %d\n", serverCmd.Process.Pid)

	return nil
}

// StopTestServer gracefully kills the server process group and cleans up resources.
func StopTestServer() {
	if serverCmd != nil && serverCmd.Process != nil {
		pid := serverCmd.Process.Pid
		fmt.Printf("  - Killing server process group (PID: %d)...\n", pid)

		pgid, err := syscall.Getpgid(pid)
		if err != nil {
			fmt.Printf("  - Warning: failed to get process group: %v\n", err)
			_ = serverCmd.Process.Kill()
		} else {
			if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
				fmt.Printf("  - Warning: failed to kill process group: %v\n", err)
				_ = serverCmd.Process.Kill()
			}
		}

		if err := serverCmd.Wait(); err != nil {
			fmt.Printf("  - Warning: server wait returned: %v\n", err)
		}

		serverCmd = nil
		fmt.Printf("  - Process group killed\n")
	}

	if serverLogFile != nil {
		serverLogFile.Close()
		serverLogFile = nil
	}

	cleanupServerBinary()
	CleanupRuntimeConfig()
	fmt.Printf("  ✅ Server stopped\n")
}

// WaitForServer polls the health endpoint until the server is ready or timeout expires.
func WaitForServer(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	healthURL := baseURL + "/api/v1/health"
	startTime := time.Now()
	lastProgressUpdate := startTime
	attemptCount := 0

	for time.Now().Before(deadline) {
		attemptCount++
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			elapsed := time.Since(startTime)
			fmt.Printf("  ✅ Server ready after %v (%d attempts)\n", elapsed.Round(time.Second), attemptCount)
			return nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		now := time.Now()
		if now.Sub(lastProgressUpdate) >= 10*time.Second {
			elapsed := now.Sub(startTime)
			remaining := deadline.Sub(now)
			fmt.Printf("  ⏳ Still waiting... (elapsed: %v, remaining: %v, attempts: %d)\n",
				elapsed.Round(time.Second), remaining.Round(time.Second), attemptCount)
			lastProgressUpdate = now
		}

		time.Sleep(200 * time.Millisecond)
	}

	elapsed := time.Since(startTime)
	return fmt.Errorf("server not ready after %v (%d attempts)", elapsed.Round(time.Second), attemptCount)
}

// GetServerLogPath returns the current server log file path.
func GetServerLogPath() string {
	return serverLogPath
}

func resolveServerBinary(projectRoot string) (string, error) {
	explicitPath := strings.TrimSpace(os.Getenv("SCAFFOLD_E2E_BINARY"))
	if explicitPath != "" {
		fmt.Printf("  - Using explicit E2E binary: %s\n", explicitPath)
		serverBinPath = ""
		return explicitPath, nil
	}

	if isCIEnvironment() {
		binaryPath := filepath.Join(projectRoot, platformBinaryName())
		if _, err := os.Stat(binaryPath); err == nil {
			fmt.Printf("  - Using CI pre-built binary: %s\n", binaryPath)
			serverBinPath = ""
			return binaryPath, nil
		}
	}

	return buildFreshServerBinary(projectRoot)
}

func buildFreshServerBinary(projectRoot string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "scaffold-e2e-bin-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir for E2E binary: %w", err)
	}

	binaryPath := filepath.Join(tmpDir, "scaffold-e2e")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./")
	buildCmd.Dir = projectRoot
	buildCmd.Env = os.Environ()

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to build fresh E2E binary: %w: %s", err, strings.TrimSpace(string(output)))
	}

	serverBinPath = binaryPath
	fmt.Printf("  - Built fresh test binary: %s\n", binaryPath)

	return binaryPath, nil
}

func cleanupServerBinary() {
	if serverBinPath == "" {
		return
	}

	if err := os.RemoveAll(filepath.Dir(serverBinPath)); err != nil {
		fmt.Printf("  - Warning: failed to remove temp binary dir: %v\n", err)
	}

	serverBinPath = ""
}

func isCIEnvironment() bool {
	return os.Getenv("CI") == "true" || os.Getenv("GITLAB_CI") == "true"
}

func platformBinaryName() string {
	if runtime.GOOS == "darwin" {
		return "scaffold-darwin"
	}

	return "scaffold-amd64"
}
