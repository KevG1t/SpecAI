//go:build !windows

package update

import "os/exec"

// configureBackgroundCmd detaches stdin so a background subprocess cannot
// consume the controlling terminal's input. On non-Windows platforms there is
// no console-mode inheritance problem, so no creation flags are needed.
func configureBackgroundCmd(cmd *exec.Cmd) {
	if cmd != nil {
		cmd.Stdin = nil
	}
}
