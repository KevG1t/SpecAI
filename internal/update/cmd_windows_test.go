//go:build windows

package update

import (
	"os/exec"
	"testing"
)

// On Windows, configureBackgroundCmd must give the child its own hidden console
// (CREATE_NO_WINDOW) so it cannot reset the parent console's input mode and
// freeze the Bubbletea TUI.
func TestConfigureBackgroundCmdSetsCreateNoWindow(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "ver")

	configureBackgroundCmd(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr is nil, want CREATE_NO_WINDOW configured")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatalf("CreationFlags = %#x, want CREATE_NO_WINDOW (%#x) bit set",
			cmd.SysProcAttr.CreationFlags, createNoWindow)
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow = false, want true")
	}
}
