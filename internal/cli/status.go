package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KevG1t/SpecAI/internal/state"
	"github.com/KevG1t/SpecAI/internal/system"
)

// RunStatus reports the currently installed agents, installed skills, SDD-Memory
// health, and detected platform profile.
func RunStatus(args []string, stdout io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	_, _ = fmt.Fprintln(stdout, "specai status")
	_, _ = fmt.Fprintln(stdout, "=============")
	_, _ = fmt.Fprintln(stdout, "")

	// Platform profile.
	result, err := system.Detect(context.Background())
	if err == nil {
		p := result.System.Profile
		_, _ = fmt.Fprintf(stdout, "Platform:    %s/%s\n", result.System.OS, result.System.Arch)
		_, _ = fmt.Fprintf(stdout, "Package mgr: %s\n", p.PackageManager)
		if p.IsWSL {
			_, _ = fmt.Fprintln(stdout, "Environment: WSL2")
		} else if p.IsTermux {
			_, _ = fmt.Fprintln(stdout, "Environment: Termux")
		}
		_, _ = fmt.Fprintln(stdout, "")
	}

	// Installed agents.
	s, err := state.Read(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(stdout, "Agents: none installed (run 'specai install')")
		} else {
			_, _ = fmt.Fprintf(stdout, "Agents: error reading state — %v\n", err)
		}
	} else if len(s.InstalledAgents) == 0 {
		_, _ = fmt.Fprintln(stdout, "Agents: none installed")
	} else {
		_, _ = fmt.Fprintf(stdout, "Agents (%d):\n", len(s.InstalledAgents))
		for _, id := range s.InstalledAgents {
			dir := system.AgentConfigDir(homeDir, id)
			marker := "ok"
			if dir != "" {
				if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
					marker = "MISSING config dir"
				}
			}
			_, _ = fmt.Fprintf(stdout, "  %s — %s\n", id, marker)
		}
	}

	_, _ = fmt.Fprintln(stdout, "")

	// SDD-Memory health.
	sddHealth := statusCheckSDDMemory()
	_, _ = fmt.Fprintf(stdout, "SDD-Memory: %s\n", sddHealth)

	return nil
}

// statusCheckSDDMemory checks whether the sdd-memory HTTP health endpoint is reachable.
func statusCheckSDDMemory() string {
	const healthEnvVar = "SDD_MEMORY_BASE_URL"
	baseURL := os.Getenv(healthEnvVar)
	if baseURL == "" {
		baseURL = "http://localhost:7437"
	}
	healthURL := strings.TrimRight(baseURL, "/") + "/health"

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthURL) //nolint:noctx
	if err != nil {
		return "degraded (unreachable at " + healthURL + ")"
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "healthy (" + healthURL + ")"
	}
	return fmt.Sprintf("degraded (HTTP %d at %s)", resp.StatusCode, healthURL)
}
