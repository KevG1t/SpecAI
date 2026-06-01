package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Version holds the current build version injected from main.
var Version = "dev"

// SetVersion sets the application version.
func SetVersion(v string) {
	Version = v
}

// GetVersion returns the current application version.
func GetVersion() string {
	return Version
}

// Start initializes the TUI and starts the Bubble Tea program.
// With the new architecture, the TUI stays alive until the user explicitly quits.
func Start() error {
	p := tea.NewProgram(NewMainModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
