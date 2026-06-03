package sddmemory

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var (
	lookPath    = exec.LookPath
	execCommand = exec.Command
)

func VerifyInstalled() error {
	if _, err := lookPath("sdd-memory"); err != nil {
		return fmt.Errorf("sdd-memory binary not found in PATH: %w", err)
	}
	return nil
}

func VerifyVersion() (string, error) {
	cmd := execCommand("sdd-memory", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("sdd-memory version command failed: %w", err)
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return "", fmt.Errorf("sdd-memory version returned empty output")
	}

	return version, nil
}

func VerifyHealth(ctx context.Context, baseURL string) error {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://127.0.0.1:7437"
	}

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/health", nil)
	if err != nil {
		return fmt.Errorf("build sdd-memory health request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sdd-memory health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sdd-memory health check returned status %d", resp.StatusCode)
	}

	return nil
}
