package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// DetectFunc is the injectable function used to run system detection.
// Using a function value instead of an interface avoids over-engineering a
// single-method boundary while still enabling test doubles.
type DetectFunc func(ctx context.Context) (*system.DetectionResult, error)

// detectionResultMsg is the internal async message carrying the detection outcome.
type detectionResultMsg struct {
	result *system.DetectionResult
	err    error
}

// DetectionConfirmedMsg is emitted when the user confirms (or acknowledges) the
// detection screen, regardless of whether detection succeeded or failed.
type DetectionConfirmedMsg struct {
	Result *system.DetectionResult
}

// DetectionModel is a standalone BubbleTea model for the system detection screen.
type DetectionModel struct {
	detectFunc DetectFunc
	result     *system.DetectionResult
	err        error
	ready      bool
}

// NewDetectionModel constructs a DetectionModel with an injected detect function.
func NewDetectionModel(fn DetectFunc) DetectionModel {
	return DetectionModel{detectFunc: fn}
}

// Init fires the async detection command.
func (m DetectionModel) Init() tea.Cmd {
	fn := m.detectFunc
	return func() tea.Msg {
		r, err := fn(context.Background())
		return detectionResultMsg{result: r, err: err}
	}
}

// Update handles detection result and user key presses.
func (m DetectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case detectionResultMsg:
		m.ready = true
		m.result = msg.result
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter", " ":
			if m.ready {
				result := m.result
				return m, func() tea.Msg {
					return DetectionConfirmedMsg{Result: result}
				}
			}
		}
	}
	return m, nil
}

// View renders the detection screen.
func (m DetectionModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("System Detection"))
	b.WriteString("\n\n")

	if !m.ready {
		b.WriteString(styles.SubtextStyle.Render("Detecting your system... (please wait)"))
		b.WriteString("\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(styles.ErrorStyle.Render("Detection error: "+m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(styles.SubtextStyle.Render("Press enter to continue anyway"))
		b.WriteString("\n")
		return b.String()
	}

	if m.result != nil {
		r := m.result
		b.WriteString(styles.HeadingStyle.Render("System"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  OS:           %s\n", r.System.OS))
		b.WriteString(fmt.Sprintf("  Arch:         %s\n", r.System.Arch))
		b.WriteString(fmt.Sprintf("  Shell:        %s\n", r.System.Shell))
		b.WriteString(fmt.Sprintf("  Supported:    %s\n", boolStr(r.System.Supported)))
		b.WriteString("\n")

		if len(r.Tools) > 0 {
			b.WriteString(styles.HeadingStyle.Render("Tools"))
			b.WriteString("\n")
			for name, status := range r.Tools {
				if status.Installed {
					b.WriteString(fmt.Sprintf("  %s: %s\n", name, styles.SuccessStyle.Render("installed")))
				} else {
					b.WriteString(fmt.Sprintf("  %s: %s\n", name, styles.SubtextStyle.Render("not found")))
				}
			}
			b.WriteString("\n")
		}

		if len(r.Configs) > 0 {
			b.WriteString(styles.HeadingStyle.Render("Agent Configs Found"))
			b.WriteString("\n")
			for _, cfg := range r.Configs {
				if cfg.Exists {
					b.WriteString(fmt.Sprintf("  %s: %s\n", cfg.Agent, styles.SuccessStyle.Render("found")))
				}
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(styles.HelpStyle.Render("enter/space: continue"))

	return b.String()
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
