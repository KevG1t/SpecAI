package update

import (
	"os"
	"os/exec"
	"testing"
)

// configureBackgroundCmd must detach stdin on every platform so a background
// probe cannot consume the controlling terminal's input.
func TestConfigureBackgroundCmdDetachesStdin(t *testing.T) {
	cmd := exec.Command("go", "version")
	cmd.Stdin = os.Stdin // simulate an inherited console stdin

	configureBackgroundCmd(cmd)

	if cmd.Stdin != nil {
		t.Fatalf("Stdin = %v, want nil after configureBackgroundCmd", cmd.Stdin)
	}
}

// configureBackgroundCmd must be safe to call with a nil command.
func TestConfigureBackgroundCmdNilSafe(t *testing.T) {
	configureBackgroundCmd(nil) // must not panic
}
