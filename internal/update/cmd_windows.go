//go:build windows

package update

import (
	"os/exec"
	"syscall"
)

// createNoWindow is the Windows CREATE_NO_WINDOW process creation flag. A child
// started with it gets its own hidden console instead of inheriting the parent's.
const createNoWindow = 0x08000000

// configureBackgroundCmd makes a subprocess safe to spawn while a Bubbletea TUI
// owns the Windows console.
//
// Without this, the child (e.g. powershell wrapping a .ps1 version probe)
// inherits the parent console and resets its input mode on start/exit. That
// destroys the raw-mode console state Bubbletea relies on, leaving the TUI
// rendering correctly but unable to read keystrokes — the "frozen welcome
// screen" symptom. CREATE_NO_WINDOW gives the child its own hidden console so it
// can no longer touch ours; detaching stdin prevents it from consuming input.
func configureBackgroundCmd(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
	cmd.Stdin = nil
}
