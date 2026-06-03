package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/state"
	"github.com/KevG1t/SpecAI/internal/system"
)

// RunBackup creates a timestamped snapshot of all installed agent configurations
// and skill files. It exits gracefully with a message when no agents are installed.
func RunBackup(args []string, stdout io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	// Read current install state to discover configured agent directories.
	s, err := state.Read(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(stdout, "Nothing to back up: no SpecAI install state found.")
			_, _ = fmt.Fprintln(stdout, "Run 'specai install' first.")
			return nil
		}
		return fmt.Errorf("read install state: %w", err)
	}

	if len(s.InstalledAgents) == 0 {
		_, _ = fmt.Fprintln(stdout, "Nothing to back up: no agents installed.")
		return nil
	}

	// Collect the agent config files and the state file itself.
	targets := backupCollectTargets(homeDir, s.InstalledAgents)
	if len(targets) == 0 {
		_, _ = fmt.Fprintln(stdout, "Nothing to back up: agent config directories not found.")
		return nil
	}

	// Create the snapshot directory.
	snapshotID := time.Now().UTC().Format("20060102150405.000000000")
	backupRoot := filepath.Join(homeDir, ".specai", "backups")
	snapshotDir := filepath.Join(backupRoot, snapshotID)

	snap := backup.NewSnapshotter()
	manifest, err := snap.Create(snapshotDir, targets)
	if err != nil {
		return fmt.Errorf("create backup snapshot: %w", err)
	}

	_, _ = fmt.Fprintf(stdout, "Backup created: %s\n", snapshotDir)
	_, _ = fmt.Fprintf(stdout, "  Files backed up: %d\n", manifest.FileCount)
	_, _ = fmt.Fprintf(stdout, "  Agents: %v\n", s.InstalledAgents)
	return nil
}

// backupCollectTargets gathers configuration file paths for all installed agents.
func backupCollectTargets(homeDir string, agentIDs []string) []string {
	out := make([]string, 0, len(agentIDs)+1)

	// Always include the state file.
	out = append(out, state.Path(homeDir))

	for _, id := range agentIDs {
		dir := backupAgentConfigDir(homeDir, id)
		if dir == "" {
			continue
		}
		// Walk the directory to collect individual files.
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				out = append(out, path)
			}
			return nil
		})
	}

	return out
}

// backupAgentConfigDir returns the primary config directory for a known agent ID.
// Returns "" for unknown agents.
func backupAgentConfigDir(homeDir, agentID string) string {
	return system.AgentConfigDir(homeDir, agentID)
}
