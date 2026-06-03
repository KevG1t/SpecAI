package app

import (
	"fmt"
	"io"
)

func printHelp(w io.Writer, version string) {
	fmt.Fprintf(w, `specai — SpecAI: Ecosystem, Frameworks, Workflows (%s)

USAGE
  specai                     Launch interactive TUI
  specai <command> [flags]

COMMANDS
  install      Configure AI coding agents on this machine
  uninstall    Remove SpecAI managed files from this machine
  sync         Sync agent configs and skills to current version
  skill-registry refresh
               Refresh .atl/skill-registry.md with cache-hit fast path
  update       Check for available updates
  upgrade      Apply updates to managed tools
  restore      Restore a config backup
  doctor       Run ecosystem health diagnostics
  version      Print version

FLAGS
  --help, -h    Show this help

Run 'specai help' for this message.
Documentation: https://github.com/KevG1t/SpecAI
`, version)
}
