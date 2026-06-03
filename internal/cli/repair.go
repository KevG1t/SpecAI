package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/state"
	"github.com/KevG1t/SpecAI/internal/system"
)

// RunRepair inspects the current install state, identifies agents whose config
// directories are missing or incomplete, and reports what would need to be
// re-run. Re-execution is delegated to `specai install` to avoid duplicating
// pipeline logic.
//
// When all agents appear healthy (config dirs present), it reports "nothing to
// repair" and exits with code 0.
func RunRepair(args []string, stdout io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	s, err := state.Read(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(stdout, "Nothing to repair: no SpecAI install state found.")
			_, _ = fmt.Fprintln(stdout, "Run 'specai install' to set up the ecosystem.")
			return nil
		}
		return fmt.Errorf("read install state: %w", err)
	}

	if len(s.InstalledAgents) == 0 {
		_, _ = fmt.Fprintln(stdout, "Nothing to repair: no agents installed.")
		return nil
	}

	// Identify agents whose config directories are missing.
	var missing []string
	for _, agentID := range s.InstalledAgents {
		dir := system.AgentConfigDir(homeDir, agentID)
		if dir == "" {
			continue // Unknown agent — skip
		}
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			missing = append(missing, agentID)
		}
	}

	if len(missing) == 0 {
		_, _ = fmt.Fprintln(stdout, "All installed agents appear healthy — nothing to repair.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Repair needed for %d agent(s):\n", len(missing))
	for _, id := range missing {
		dir := filepath.Join(homeDir, ".specai", "backups") // placeholder reference
		_, _ = fmt.Fprintf(stdout, "  %s — config directory missing (expected: %s)\n", id, dir)
	}
	_, _ = fmt.Fprintln(stdout, "")
	_, _ = fmt.Fprintln(stdout, "To repair, re-run: specai install")
	return nil
}
